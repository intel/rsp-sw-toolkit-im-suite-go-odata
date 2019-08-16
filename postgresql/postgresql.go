package postgresql

import (
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
	"eq": "=",
	"nq": "!=",
	"gt": ">",
	"ge": ">=",
	"lt": "<",
	"le": "<=",
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

	// Limit & Offset
	limit, existLimit := queryMap[parser.Top].(int)
	skip, existSkip := queryMap[parser.Skip].(int)

	if existLimit {
		finalQuery.WriteString(" LIMIT ")
		finalQuery.WriteString(strconv.Itoa(limit))
		finalQuery.WriteString(" ")
	}

	if existSkip {
		finalQuery.WriteString(" OFFSET ")
		finalQuery.WriteString(strconv.Itoa(skip))
		finalQuery.WriteString(" ")
	}

	queryString := finalQuery.String()
	rows, err := db.NamedQuery(queryString, valuesMap)
	if err != nil {
		return nil, err
	}

	return rows, nil

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

				if selectSlice.Len() > 1 {
					selectClause.WriteString(",")
				}
			}
		}
	} else {
		selectClause.WriteString(" * ")
	}

	return selectClause.String()

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

			if operator, ok := sqlOperators[node.Token.Value.(string)]; ok {
				filter.WriteString(operator)
			} else {
				return "", ErrInvalidInput
			}
			filter.WriteString(":")
			filter.WriteString(node.Children[0].Token.Value.(string))
			valuesMap[node.Children[0].Token.Value.(string)] = node.Children[0].Token.Value.(string)

		case "and":
			leftFilter, err := applyFilter(node.Children[0], column) // Left children
			if err != nil {
				return "", err
			}
			rightFilter, _ := applyFilter(node.Children[1], column) // Right children
			if err != nil {
				return "", err
			}

			filter.WriteString(leftFilter)
			filter.WriteString(" AND ")
			filter.WriteString(rightFilter)

		case "or":
			leftFilter, err := applyFilter(node.Children[0], column) // Left children
			if err != nil {
				return "", err
			}
			rightFilter, err := applyFilter(node.Children[1], column) // Right children
			if err != nil {
				return "", err
			}
			filter.WriteString(leftFilter)
			filter.WriteString(" OR ")
			filter.WriteString(rightFilter)

			// //Functions
			// case "startswith":
			// 	if _, ok := node.Children[1].Token.Value.(string); !ok {
			// 		return nil, ErrInvalidInput
			// 	}
			// 	node.Children[1].Token.Value = strings.Replace(node.Children[1].Token.Value.(string), "'", "", -1)
			// 	//nolint: vet
			// 	value := bson.RegEx{"^" + node.Children[1].Token.Value.(string), "gi"}
			// 	filter[node.Children[0].Token.Value.(string)] = value

			// case "endswith":
			// 	if _, ok := node.Children[1].Token.Value.(string); !ok {
			// 		return nil, ErrInvalidInput
			// 	}
			// 	node.Children[1].Token.Value = strings.Replace(node.Children[1].Token.Value.(string), "'", "", -1)
			// 	//nolint: vet
			// 	value := bson.RegEx{node.Children[1].Token.Value.(string) + "$", "gi"}
			// 	filter[node.Children[0].Token.Value.(string)] = value

			// case "contains":
			// 	if _, ok := node.Children[1].Token.Value.(string); !ok {
			// 		return nil, ErrInvalidInput
			// 	}
			// 	node.Children[1].Token.Value = strings.Replace(node.Children[1].Token.Value.(string), "'", "", -1)
			// 	//nolint: vet
			// 	value := bson.RegEx{node.Children[1].Token.Value.(string), "gi"}
			// 	filter[node.Children[0].Token.Value.(string)] = value

		}
	}
	return filter.String(), nil
}
