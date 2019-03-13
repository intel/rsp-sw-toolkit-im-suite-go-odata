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

package parser

import (
	"errors"
	"net/url"
	"strings"
)

// Odata keywords
const (
	Select      = "$select"
	Top         = "$top"
	Skip        = "$skip"
	Count       = "$count"
	OrderBy     = "$orderby"
	InlineCount = "$inlinecount"
	Filter      = "$filter"
)

// ParseURLValues parses url values in odata format into a map of interfaces for the DB adapters to translate
//nolint :gocyclo
func ParseURLValues(query url.Values) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	var parseErrors []string

	result[Count] = false
	result[InlineCount] = "none"

	if isCountAndInlineCountSet(query) {
		parseErrors = append(parseErrors, "$count and $inlinecount cannot be set in the same odata query")
	}

	for queryParam, queryValues := range query {
		var parseResult interface{}
		var err error

		if len(queryValues) > 1 {
			parseErrors = append(parseErrors, "Duplicate keyword '"+queryParam+"' found in odata query")
			continue
		}
		value := query.Get(queryParam)
		if value == "" && queryParam != Count {
			parseErrors = append(parseErrors, "No value was set for keyword '"+queryParam+"'")
			continue
		}

		switch queryParam {
		case Select:
			parseResult, err = parseStringArray(&value)
		case Top:
			parseResult, err = parseInt(&value)
		case Skip:
			parseResult, err = parseInt(&value)
		case Count:
			parseResult = true
		case OrderBy:
			parseResult, err = parseOrderArray(&value)
		case InlineCount:
			if !isValidInlineCountValue(value) {
				parseErrors = append(parseErrors, "Inline count value needs to be allpages or none")
			}
			parseResult = value
		case Filter:
			parseResult, err = parseFilterString(value)
		default:
			parseErrors = append(parseErrors, "Keyword '"+queryParam+"' is not valid")
		}

		if err != nil {
			parseErrors = append(parseErrors, err.Error())
		}
		result[queryParam] = parseResult
	}
	if len(parseErrors) > 0 {
		return nil, errors.New(strings.Join(parseErrors[:], ";"))
	}
	return result, nil
}

func isValidInlineCountValue(value string) bool {
	valueNoSpace := strings.TrimSpace(value)
	if valueNoSpace != "allpages" && valueNoSpace != "none" {
		return false
	}
	return true
}

func isCountAndInlineCountSet(query url.Values) bool {

	_, countFound := query[Count]
	_, inlineFound := query[InlineCount]

	if countFound && inlineFound {
		return true
	}

	return false
}
