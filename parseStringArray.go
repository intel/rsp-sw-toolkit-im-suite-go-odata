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
	"strings"

	"github.impcloud.net/RSP-Inventory-Suite/go-odata/validatefield"
)

// OrderItem holds order key information
type OrderItem struct {
	Field string
	Order string
}

func parseStringArray(value *string) ([]string, error) {
	result := strings.Split(*value, ",")

	// trim out space
	for idx, resultNoSpace := range result {
		result[idx] = strings.TrimSpace(resultNoSpace)
	}

	if len(result) == 0 {
		return nil, errors.New("cannot parse zero length string")
	}

	return result, nil
}

func parseOrderArray(value *string) ([]OrderItem, error) {
	parsedArray, err := parseStringArray(value)
	if err != nil {
		return nil, err
	}

	// Validate values for special characters
	valid := validatefield.New("~!@#$%^&*()_+-")
	for _, val := range parsedArray {
		if valid.ValidateField(val) || val == "" {
			return nil, errors.New("Cannot support field " + val)
		}
	}

	result := make([]OrderItem, len(parsedArray))

	for i, v := range parsedArray {
		timmedString := strings.TrimSpace(v)
		compressedSpaces := strings.Join(strings.Fields(timmedString), " ")
		s := strings.Split(compressedSpaces, " ")

		if len(s) > 2 {
			return nil, errors.New("Cannot have more than 2 items in orderby query")
		}

		if len(s) > 1 {
			if s[1] != "asc" &&
				s[1] != "desc" {
				return nil, errors.New("Second value in orderby needs to be asc or desc")
			}
			result[i] = OrderItem{s[0], s[1]}
			continue
		}
		result[i] = OrderItem{compressedSpaces, "asc"}
	}
	return result, nil
}
