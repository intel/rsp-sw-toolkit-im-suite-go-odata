package postgresql

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
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
	"contains":   "%%s%",
	"endswith":   "%%s",
	"startswith": "%s%",
}

var valuesMap = make(map[string]interface{})

// ODataSQLQuery builds a SQL like query based on OData 2.0 specification
func ODataSQLQuery(query url.Values, table string, column string, db *sqlx.DB) (*sqlx.Rows, error) {

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
	finalQuery.WriteString(table)

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

	rows, err := db.NamedQuery(finalQuery.String(), valuesMap)
	if err != nil {
		return nil, err
	}
	return rows, nil

}

// ODataCount returns the number of rows from a table
func ODataCount(db *sqlx.DB, table string) (int, error) {
	var count int
	var query strings.Builder

	query.WriteString("SELECT count(*) FROM ")
	query.WriteString(table)

	if err := db.Get(&count, query.String()); err != nil {
		return 0, err
	}
	return count, nil
}

func buildSelectClause(queryMap map[string]interface{}, column string) string {

	// Select clause
	// 'data' is the column name of the jsonb data
	var selectClause strings.Builder
	selectClause.WriteString("SELECT ")
	if queryMap["$select"] != nil {
		selectSlice := reflect.ValueOf(queryMap["$select"])
		if selectSlice.Len() > 1 && selectSlice.Index(0).Interface().(string) != "*" {
			for i := 0; i < selectSlice.Len(); i++ {
				fieldName := selectSlice.Index(i).Interface().(string)
				selectClause.WriteString(column)
				selectClause.WriteString(" -> ")
				selectClause.WriteString("'")
				selectClause.WriteString(fieldName)
				selectClause.WriteString("'")

				selectClause.WriteString(" AS ")
				selectClause.WriteString(fieldName)

				if selectSlice.Len() > i+1 {
					selectClause.WriteString(",")
				}
			}
		}
	} else {
		selectClause.WriteString(" * ")
	}

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

	if queryMap[parser.OrderBy] != nil {
		orderBySlice := queryMap[parser.OrderBy].([]parser.OrderItem)
		for id, item := range orderBySlice {
			query.WriteString(column)
			query.WriteString(" ->> ")
			query.WriteString("'")
			query.WriteString(item.Field)
			query.WriteString("'")
			if item.Order == "desc" {
				query.WriteString(" DESC ")
			}

			if len(orderBySlice) > id+1 {
				query.WriteString(",")
			}
		}
	}

	return query.String()
}

func applyFilter(node *parser.ParseNode, column string) (string, error) {

	var filter strings.Builder

	if _, ok := node.Token.Value.(string); ok {

		switch node.Token.Value {

		case "eq", "ne", "gt", "ge", "lt", "le":
			if _, keyOk := node.Children[0].Token.Value.(string); !keyOk {
				return "", ErrInvalidInput
			}
			filter.WriteString(column)
			filter.WriteString(" ->> ")
			filter.WriteString("'")
			filter.WriteString(node.Children[0].Token.Value.(string))
			filter.WriteString("'")

			operator, _ := sqlOperators[node.Token.Value.(string)]
			filter.WriteString(operator)

			filter.WriteString(":")
			filter.WriteString(node.Children[0].Token.Value.(string))
			valuesMap[node.Children[0].Token.Value.(string)] = node.Children[1].Token.Value

		case "or", "and":
			leftFilter, err := applyFilter(node.Children[0], column) // Left children
			if err != nil {
				return "", err
			}
			rightFilter, err := applyFilter(node.Children[1], column) // Right children
			if err != nil {
				return "", err
			}
			filter.WriteString(leftFilter)
			filter.WriteString(" ")
			filter.WriteString(node.Token.Value.(string)) // AND/OR
			filter.WriteString(" ")
			filter.WriteString(rightFilter)

		//Functions
		case "contains", "endswith", "startswith":
			if _, ok := node.Children[1].Token.Value.(string); !ok {
				return "", ErrInvalidInput
			}
			// Remove single quote
			node.Children[1].Token.Value = strings.Replace(node.Children[1].Token.Value.(string), "'", "", -1)
			filter.WriteString(column)
			filter.WriteString(" ->> ")
			filter.WriteString("'")
			filter.WriteString(node.Children[0].Token.Value.(string))
			filter.WriteString("'")
			filter.WriteString(" LIKE ")
			operator, _ := sqlOperators[node.Token.Value.(string)]
			result := fmt.Sprintf(operator, node.Children[1].Token.Value.(string))
			filter.WriteString("'")
			filter.WriteString(result)
			filter.WriteString("'")
		}
	}
	return filter.String(), nil
}
