package postgresql

import (
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.impcloud.net/RSP-Inventory-Suite/go-odata/parser"
)

// ErrInvalidInput Client errors
var ErrInvalidInput = errors.New("odata syntax error")

var sqlOperators = map[string]string{
	"eq":         "=",
	"ne":         "!=",
	"gt":         ">",
	"ge":         ">=",
	"lt":         "<",
	"le":         "<=",
	"or":         "or",
	"and":        "and",
	"contains":   "%%%s%%",
	"endswith":   "%%%s",
	"startswith": "%s%%",
}

// ODataSQLQuery builds a SQL like query based on OData 2.0 specification
func ODataSQLQuery(query url.Values, table string, column string, db *sql.DB) (*sql.Rows, error) {

	// Parse url values
	queryMap, err := parser.ParseURLValues(query)
	if err != nil {
		return nil, errors.Wrap(ErrInvalidInput, err.Error())
	}

	var finalQuery strings.Builder

	// SELECT clause
	finalQuery.WriteString(buildSelectClause(queryMap, column))

	// FROM clause
	finalQuery.WriteString(" FROM ")
	finalQuery.WriteString(pq.QuoteIdentifier(table))

	// WHERE clause
	if queryMap[parser.Filter] != nil {
		finalQuery.WriteString(" WHERE ")
		filterQuery, _ := queryMap[parser.Filter].(*parser.ParseNode)
		filterClause, err := applyFilter(filterQuery, column)
		if err != nil {
			return nil, errors.Wrap(ErrInvalidInput, err.Error())
		}

		finalQuery.WriteString(filterClause)
	}

	// Order by
	if queryMap[parser.OrderBy] != nil {
		finalQuery.WriteString(buildOrderBy(queryMap, column))
	}

	// Limit & Offset
	finalQuery.WriteString(buildLimitSkipClause(queryMap))

	rows, err := db.Query(finalQuery.String())
	if err != nil {
		return nil, err
	}
	return rows, nil

}

// ODataCount returns the number of rows from a table
func ODataCount(db *sql.DB, table string) (int, error) {
	var count int
	selectStmt := fmt.Sprintf("SELECT count(*) FROM %s", pq.QuoteIdentifier(table))
	row := db.QueryRow(selectStmt)
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func buildSelectClause(queryMap map[string]interface{}, column string) string {

	// Select clause
	// 'data' is the column name of the jsonb data
	selectSlice, _ := queryMap["$select"].([]string)
	if len(selectSlice) == 0 {
		return "SELECT * "
	}

	var selectClause strings.Builder
	selectClause.WriteString("SELECT ")
	col := pq.QuoteIdentifier(column)

	for _, fieldName := range selectSlice[:len(selectSlice)-1] {
		fmt.Fprintf(&selectClause, "%s -> %s AS %s, ",
			col, pq.QuoteLiteral(fieldName), pq.QuoteIdentifier(fieldName))
	}

	// last one without a comma
	fieldName := selectSlice[len(selectSlice)-1]
	fmt.Fprintf(&selectClause, "%s -> %s AS %s ", col, pq.QuoteLiteral(fieldName), pq.QuoteIdentifier(fieldName))

	return selectClause.String()
}

func buildLimitSkipClause(queryMap map[string]interface{}) string {

	limit, existLimit := queryMap[parser.Top].(int)
	skip, existSkip := queryMap[parser.Skip].(int)

	var queryString strings.Builder

	if existLimit {
		queryString.WriteString(" LIMIT ")
		queryString.WriteString(strconv.Itoa(limit))
	}

	if existSkip {
		queryString.WriteString(" OFFSET ")
		queryString.WriteString(strconv.Itoa(skip))
	}

	return queryString.String()

}

func buildOrderBy(queryMap map[string]interface{}, column string) string {

	var query strings.Builder
	query.WriteString(" ORDER BY ")

	col := pq.QuoteIdentifier(column)
	orderBySlice := queryMap[parser.OrderBy].([]parser.OrderItem)

	for id, item := range orderBySlice {
		fmt.Fprintf(&query, "%s ->> '%s'", col, pq.QuoteIdentifier(item.Field))
		if item.Order == "desc" {
			query.WriteString(" DESC ")
		}

		if len(orderBySlice) > id+1 {
			query.WriteString(",")
		}
	}

	return query.String()
}

func applyFilter(node *parser.ParseNode, column string) (string, error) {

	if len(node.Children) != 2 {
		return "", ErrInvalidInput
	}

	var filter strings.Builder

	operator := node.Token.Value.(string)
	sqlOp := sqlOperators[operator]
	if operator == "" || sqlOp == "" {
		// invalid or unknown operator
		return "", ErrInvalidInput
	}

	switch operator {

	case "eq", "ne", "gt", "ge", "lt", "le":

		if _, keyOk := node.Children[0].Token.Value.(string); !keyOk {
			return "", ErrInvalidInput
		}

		left := pq.QuoteLiteral(node.Children[0].Token.Value.(string))

		if value, valueOk := node.Children[1].Token.Value.(string); valueOk {
			node.Children[1].Token.Value = escapeQuote(value)
		}

		right := pq.QuoteLiteral(fmt.Sprintf("%v", node.Children[1].Token.Value))

		fmt.Fprintf(&filter, "%s ->> %s %s %s", pq.QuoteIdentifier(column), left, sqlOp, right)

	case "or", "and":

		leftFilter, err := applyFilter(node.Children[0], column) // Left children
		if err != nil {
			return "", err
		}
		rightFilter, err := applyFilter(node.Children[1], column) // Right children
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&filter, "%s %s %s", leftFilter, operator, rightFilter)

	//Functions
	case "contains", "endswith", "startswith":
		if _, ok := node.Children[1].Token.Value.(string); !ok {
			return "", ErrInvalidInput
		}
		// Remove single quote
		value, valueOk := node.Children[1].Token.Value.(string)
		if !valueOk {
			return "", ErrInvalidInput
		}
		node.Children[1].Token.Value = escapeQuote(value)

		left := pq.QuoteLiteral(node.Children[0].Token.Value.(string))
		right := pq.QuoteLiteral(fmt.Sprintf(sqlOp, node.Children[1].Token.Value.(string)))

		fmt.Fprintf(&filter, "%s ->> %s LIKE %s", pq.QuoteIdentifier(column), left, right)
	}

	return filter.String(), nil
}

func escapeQuote(value string) string {

	if len(value) < 1 {
		return ""
	}

	if value[0] == '\'' {
		value = value[1:]
	}
	if value[len(value)-1] == '\'' {
		value = value[:len(value)-1]
	}

	return value
}
