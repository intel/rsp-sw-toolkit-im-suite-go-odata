/*
 * INTEL CONFIDENTIAL
 * Copyright (2017) Intel Corporation.
 *
 * The source code contained or described herein and all documents related to the source code ("Material")
 * are owned by Intel Corporation or its suppliers or licensors. Title to the Material remains with
 * Intel Corporation or its suppliers and licensors. The Material may contain trade secrets and proprietary
 * and confidential information of Intel Corporation and its suppliers and licensors, and is protected by
 * worldwide copyright and trade secret laws and treaty provisions. No part of the Material may be used,
 * copied, reproduced, modified, published, uploaded, posted, transmitted, distributed, or disclosed in
 * any way without Intel/'s prior express written permission.
 * No license under any patent, copyright, trade secret or other intellectual property right is granted
 * to or conferred upon you by disclosure or delivery of the Materials, either expressly, by implication,
 * inducement, estoppel or otherwise. Any license under such intellectual property rights must be express
 * and approved by Intel in writing.
 * Unless otherwise agreed by Intel in writing, you may not remove or alter this notice or any other
 * notice embedded in Materials by Intel or Intel's suppliers or licensors in any way.
 */

package mongo

import (
	"encoding/hex"
	"net/url"
	"reflect"
	"strings"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/pkg/errors"
	"github.impcloud.net/RSP-Inventory-Suite/go-odata/parser"
)

// ErrInvalidInput Client errors
var ErrInvalidInput = errors.New("odata syntax error")

// Making FilterObj global for now, so inlinecount can reuse it.
// TODO: Refactor odata library API
var filterObj bson.M

// ODataQuery creates a mgo query based on odata parameters
//nolint :gocyclo
func ODataQuery(query url.Values, object interface{}, collection *mgo.Collection) error {

	// Parse url values
	queryMap, err := parser.ParseURLValues(query)
	if err != nil {
		return errors.Wrap(ErrInvalidInput, err.Error())
	}

	limit, _ := queryMap[parser.Top].(int)
	skip, _ := queryMap[parser.Skip].(int)

	filterObj = make(bson.M)
	if queryMap[parser.Filter] != nil {
		filterQuery, _ := queryMap[parser.Filter].(*parser.ParseNode)
		var err error
		filterObj, err = applyFilter(filterQuery)
		if err != nil {
			return errors.Wrap(ErrInvalidInput, err.Error())
		}
	}

	// Prepare Select
	selectMap := make(bson.M)

	if queryMap["$select"] != nil {
		selectSlice := reflect.ValueOf(queryMap["$select"])
		if selectSlice.Len() == 1 && selectSlice.Index(0).Interface().(string) == "*" {
			// Do nothing, same as no $select
		} else {
			for i := 0; i < selectSlice.Len(); i++ {
				fieldName := selectSlice.Index(i).Interface().(string)
				selectMap[fieldName] = 1
			}
		}
	}

	// Sort
	var sortFields []string
	if queryMap[parser.OrderBy] != nil {
		orderBySlice := queryMap[parser.OrderBy].([]parser.OrderItem)
		for _, item := range orderBySlice {
			if item.Order == "desc" {
				item.Field = "-" + item.Field
			}
			sortFields = append(sortFields, item.Field)
		}
	}

	// Query
	odataFunc := collection.Find(filterObj).Select(selectMap).Limit(limit).Skip(skip).Sort(sortFields...).All(object)

	return odataFunc
}

// ODataCount runs a collection.Count() function based on $count odata parameter
func ODataCount(collection *mgo.Collection) (int, error) {
	return collection.Count()
}

// ODataInlineCount retrieves the total count from a filtered data
func ODataInlineCount(collection *mgo.Collection) (int, error) {

	return collection.Find(filterObj).Count()
}

//nolint :gocyclo
func applyFilter(node *parser.ParseNode) (bson.M, error) {

	filter := make(bson.M)

	if _, ok := node.Token.Value.(string); ok {
		switch node.Token.Value {

		case "eq":
			// Escape single quotes in the case of strings
			if _, valueOk := node.Children[1].Token.Value.(string); valueOk {
				node.Children[1].Token.Value = strings.Replace(node.Children[1].Token.Value.(string), "'", "", -1)
			}
			value := bson.M{"$" + node.Token.Value.(string): node.Children[1].Token.Value}
			if _, keyOk := node.Children[0].Token.Value.(string); !keyOk {
				return nil, ErrInvalidInput
			}
			filter[node.Children[0].Token.Value.(string)] = value

		case "ne":
			// Escape single quotes in the case of strings
			if _, valueOk := node.Children[1].Token.Value.(string); valueOk {
				node.Children[1].Token.Value = strings.Replace(node.Children[1].Token.Value.(string), "'", "", -1)
			}
			value := bson.M{"$" + node.Token.Value.(string): node.Children[1].Token.Value}
			if _, keyOk := node.Children[0].Token.Value.(string); !keyOk {
				return nil, ErrInvalidInput
			}
			filter[node.Children[0].Token.Value.(string)] = value

		case "gt":
			var keyString string
			if keyString, ok = node.Children[0].Token.Value.(string); !ok {
				return nil, ErrInvalidInput
			}

			var value bson.M
			if keyString == "_id" {
				var idString string
				if _, ok := node.Children[1].Token.Value.(string); ok {
					idString = strings.Replace(node.Children[1].Token.Value.(string), "'", "", -1)
				}
				decodedString, err := hex.DecodeString(idString)
				if err != nil || len(decodedString) != 12 {
					return nil, ErrInvalidInput
				}
				value = bson.M{"$" + node.Token.Value.(string): bson.ObjectId(decodedString)}
			} else {
				value = bson.M{"$" + node.Token.Value.(string): node.Children[1].Token.Value}
			}
			filter[keyString] = value

		case "ge":
			value := bson.M{"$gte": node.Children[1].Token.Value}
			if _, ok := node.Children[0].Token.Value.(string); !ok {
				return nil, ErrInvalidInput
			}
			filter[node.Children[0].Token.Value.(string)] = value

		case "lt":
			value := bson.M{"$" + node.Token.Value.(string): node.Children[1].Token.Value}
			if _, ok := node.Children[0].Token.Value.(string); !ok {
				return nil, ErrInvalidInput
			}
			filter[node.Children[0].Token.Value.(string)] = value

		case "le":
			value := bson.M{"$lte": node.Children[1].Token.Value}
			if _, ok := node.Children[0].Token.Value.(string); !ok {
				return nil, ErrInvalidInput
			}
			filter[node.Children[0].Token.Value.(string)] = value

		case "and":
			leftFilter, err := applyFilter(node.Children[0]) // Left children
			if err != nil {
				return nil, err
			}
			rightFilter, _ := applyFilter(node.Children[1]) // Right children
			if err != nil {
				return nil, err
			}
			filter["$and"] = []bson.M{leftFilter, rightFilter}

		case "or":
			leftFilter, err := applyFilter(node.Children[0]) // Left children
			if err != nil {
				return nil, err
			}
			rightFilter, err := applyFilter(node.Children[1]) // Right children
			if err != nil {
				return nil, err
			}
			filter["$or"] = []bson.M{leftFilter, rightFilter}

		//Functions
		case "startswith":
			if _, ok := node.Children[1].Token.Value.(string); !ok {
				return nil, ErrInvalidInput
			}
			node.Children[1].Token.Value = strings.Replace(node.Children[1].Token.Value.(string), "'", "", -1)
			//nolint: vet
			value := bson.RegEx{"^" + node.Children[1].Token.Value.(string), "gi"}
			filter[node.Children[0].Token.Value.(string)] = value

		case "endswith":
			if _, ok := node.Children[1].Token.Value.(string); !ok {
				return nil, ErrInvalidInput
			}
			node.Children[1].Token.Value = strings.Replace(node.Children[1].Token.Value.(string), "'", "", -1)
			//nolint: vet
			value := bson.RegEx{node.Children[1].Token.Value.(string) + "$", "gi"}
			filter[node.Children[0].Token.Value.(string)] = value

		case "contains":
			if _, ok := node.Children[1].Token.Value.(string); !ok {
				return nil, ErrInvalidInput
			}
			node.Children[1].Token.Value = strings.Replace(node.Children[1].Token.Value.(string), "'", "", -1)
			//nolint: vet
			value := bson.RegEx{node.Children[1].Token.Value.(string), "gi"}
			filter[node.Children[0].Token.Value.(string)] = value

		}
	}
	return filter, nil
}
