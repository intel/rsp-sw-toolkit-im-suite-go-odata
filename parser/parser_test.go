/* Apache v2 license
*  Copyright (C) <2019> Intel Corporation
*
*  SPDX-License-Identifier: Apache-2.0
 */

package parser

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"testing"
)

func TestParseTop(t *testing.T) {

	var topTests = []struct {
		input    string
		expected int
		Err      error
	}{
		{"10", 10, nil},
		{"-10", -10, nil},
		{"0", 0, nil},
		{"A", 0, strconv.ErrSyntax},
		{"", 0, errors.New("")},
	}

	for _, expectedVal := range topTests {
		var testURLString = fmt.Sprintf("http://localhost/test?$top=%s", expectedVal.input)
		testURL, err := url.Parse(testURLString)
		if err != nil {
			t.Fatalf("Failed to parse test url")
			return
		}

		parsedResult, err := ParseURLValues(testURL.Query())

		if err != nil {
			// an error happened, is it expected?
			if expectedVal.Err != nil {
				if numError, ok := err.(*strconv.NumError); ok {
					if numError.Err != expectedVal.Err {
						t.Fatalf("Expected Error: %s \tGot: %s", expectedVal.Err, numError.Err)
					} // else the error matched, so return
				}
				if _, ok := err.(error); !ok {
					t.Errorf("Failed to catch error")
				}
			}
			// else we are expecting that no error has occured
		} else {
			// else there is no error, this is ok
			if parsedResult[Top] != expectedVal.expected {
				t.Errorf("Expected: %d \tGot: %d", expectedVal.expected, parsedResult[Top])
			}
		}
	}
}

func TestParseSkip(t *testing.T) {

	var skipTests = []struct {
		input    string
		expected int
		Err      error
	}{
		{"10", 10, nil},
		{"-10", -10, nil},
		{"0", 0, nil},
		{"A", 0, strconv.ErrSyntax},
		{"", 0, errors.New("")},
	}

	for _, expectedVal := range skipTests {
		var testURLString = fmt.Sprintf("http://localhost/test?$skip=%s", expectedVal.input)
		testURL, err := url.Parse(testURLString)
		if err != nil {
			t.Fatalf("Failed to parse test url")
			return
		}

		parsedResult, err := ParseURLValues(testURL.Query())

		if err != nil {
			// an error happened, is it expected?
			if expectedVal.Err != nil {
				if numError, ok := err.(*strconv.NumError); ok {
					if numError.Err != expectedVal.Err {
						t.Fatalf("Expected Error: %s \tGot: %s", expectedVal.Err, numError.Err)
					} // else the error matched, so return
				}
				if _, ok := err.(error); !ok {
					t.Errorf("Failed to catch error")
				}
			}
			// else we are expecting that no error has occured
		} else {
			// else there is no error, this is ok
			if parsedResult[Skip] != expectedVal.expected {
				t.Errorf("Expected: %d \tGot: %d", expectedVal.expected, parsedResult[Skip])
			}
		}
	}
}

//nolint :gocyclo
func TestParseWithCount(t *testing.T) {
	testURL, err := url.Parse("http://localhost/test?$top=10&$skip=5&$select=name,age&$count&$orderby=name   asc,  age")
	if err != nil {
		t.Error("failed to parse test url")
	}

	result, err := ParseURLValues(testURL.Query())
	if err != nil {
		t.Error("Failed to parse")
	}

	if result[Top] != 10 {
		t.Error("top value not equal to 10")
	}

	if result[Skip] != 5 {
		t.Error("skip value not equal to 5")
	}

	if len(result[Select].([]string)) != 2 {
		t.Error("Select array length does not match parameters")
	}

	for i, v := range result[Select].([]string) {
		if i == 0 {
			if v != "name" {
				t.Error("select value name not found")
			}
		}
		if i == 1 {
			if v != "age" {
				t.Error("select value age not found")
			}
		}
	}

	if result[Count] != true {
		t.Error("count value not equal to true")
	}

	if len(result[OrderBy].([]OrderItem)) != 2 {
		t.Error("OrderItem array length does not match parameters")
	}

	for i, v := range result[OrderBy].([]OrderItem) {
		if i == 0 {
			if v.Field != "name" &&
				v.Order != "desc" {
				t.Error("orderby value is incorrect")
			}
		}
		if i == 1 {
			if v.Field != "age" &&
				v.Order != "asc" {
				t.Error("orderby value is incorrect")
			}
		}
	}
}

func TestParseWithDuplicates(t *testing.T) {
	testURL, err := url.Parse("http://localhost/test?$top=10&$top=5")
	if err != nil {
		t.Error("failed to parse test url")
	}

	result, err := ParseURLValues(testURL.Query())
	if err == nil {
		t.Error("Failed to catch parse duplicate")
	}

	if result != nil {
		t.Error("Failed to catch parse duplicate")
	}
}

func TestParseWithTop(t *testing.T) {
	testURL, err := url.Parse("http://localhost/test?$top=  10")
	if err != nil {
		t.Error("failed to parse test url")
	}

	result, err := ParseURLValues(testURL.Query())
	if err != nil {
		t.Error("Failed to parse query string value top")
	}
	if result != nil {
		top := result[Top].(int)
		if top != 10 {
			t.Errorf("Expected value = 10 but found top value is %d ", top)
		}
	}
}

func TestParseWithoutCount(t *testing.T) {
	testURL, err := url.Parse("http://localhost/test?$select=name,age")
	if err != nil {
		t.Error("failed to parse test url")
	}

	result, err := ParseURLValues(testURL.Query())
	if err != nil {
		t.Error("Failed to parse")
	}

	for i, v := range result[Select].([]string) {
		if i == 0 {
			if v != "name" {
				t.Error("select value name not found")
			}
		}
		if i == 1 {
			if v != "age" {
				t.Error("select value age not found")
			}
		}
	}

	if result[Count] != false {
		t.Error("count value not equal to true")
	}

	if result[InlineCount] != "none" {
		t.Error("count value not equal to true")
	}
}

func TestParseSelectWithSpace(t *testing.T) {
	testURL, err := url.Parse("http://localhost/test?$select=name       , age ")
	if err != nil {
		t.Error("failed to parse test url")
	}

	result, err := ParseURLValues(testURL.Query())
	if err != nil {
		t.Error("Failed to parse")
	}

	for i, v := range result[Select].([]string) {
		if i == 0 {
			if v != "name" {
				t.Error("select value name not found")
			}
		}
		if i == 1 {
			if v != "age" {
				t.Error("select value age not found")
			}
		}
	}
}

func TestParseWithInvalidIntValues(t *testing.T) {
	testURL, err := url.Parse("http://localhost/test?$top=top")
	if err != nil {
		t.Error("failed to parse test url")
	}

	result, err := ParseURLValues(testURL.Query())
	if err == nil {
		t.Error("Failed to catch error")
	}

	if result != nil {
		t.Error("Failed to catch error")
	}
}

func TestParseWithInvalidSkipValue(t *testing.T) {
	testURL, err := url.Parse("http://localhost/test?$top=5&$skip=skip")
	if err != nil {
		t.Error("failed to parse test url")
	}

	result, err := ParseURLValues(testURL.Query())
	if err == nil {
		t.Error("Failed to catch error")
	}

	if result != nil {
		t.Error("Failed to catch error")
	}
}

func TestParseWithInvalidOrderByvalues(t *testing.T) {
	testURL, err := url.Parse("http://localhost/test?$orderby=last name desc")
	if err != nil {
		t.Error("failed to parse test url")
	}

	result, err := ParseURLValues(testURL.Query())
	if err == nil {
		t.Error("Failed to catch error")
	}

	if result != nil {
		t.Error("Failed to catch error")
	}

	testURL2, err2 := url.Parse("http://localhost/test?$orderby=last name")
	if err2 != nil {
		t.Error("failed to parse test url")
	}

	result2, err2 := ParseURLValues(testURL2.Query())
	if err2 == nil {
		t.Error("Failed to catch error")
	}

	if result2 != nil {
		t.Error("Failed to catch error")
	}
}

func TestParseWithInvalidInlinecountValue(t *testing.T) {
	testURL, err := url.Parse("http://localhost/test?$top=5&$inlinecount=true")
	if err != nil {
		t.Error("failed to parse test url")
	}

	result, err := ParseURLValues(testURL.Query())
	if err == nil {
		t.Error("Failed to catch error")
	}

	if result != nil {
		t.Error("Failed to catch error")
	}
}

func TestParseInlinecountValueWithSpace(t *testing.T) {
	testURL, err := url.Parse("http://localhost/test?$inlinecount=  allpages")
	if err != nil {
		t.Error("failed to parse test url")
	}

	result, err := ParseURLValues(testURL.Query())
	if err != nil {
		t.Error(err)
	}

	if result == nil {
		t.Error("Failed to parse inlinecount")
	}
}

func TestParseInlinecountAndCountNegative(t *testing.T) {
	testURL, err := url.Parse("http://localhost/test?$inlinecount=allpages&$count")
	if err != nil {
		t.Error("failed to parse test url")
	}

	result, err := ParseURLValues(testURL.Query())

	if err == nil {
		t.Error("expecting an error, $count and $inlinecount cannot be in the same query", err)
	}

	if result != nil {
		t.Error("results should be nil")
	}
}

func TestParseWithInvalidKey(t *testing.T) {
	testURL, err := url.Parse("http://localhost/test?$invalidKey='no'")
	if err != nil {
		t.Error("failed to parse test url")
	}

	result, err := ParseURLValues(testURL.Query())
	if err == nil {
		t.Error("Failed to catch error")
	}

	if result != nil {
		t.Error("Failed to catch error")
	}
}

func TestParseWithFilter(t *testing.T) {
	var filterTests = []struct {
		input    string
		expected bool
		Err      error
	}{
		{"((epc_item_type gt 0) and (event ne 'departed') and (required eq true) and (count lt 0.1) or " +
			"(SKU eq '123') and contains(epc_time, 0) and startswith(epc_code, '456') or " +
			"endswith(upc_code, '789'))", true, nil}, // large valid filter case
		{"_id gt '59a6fbaf22e60174f5107a9a' and upc_code eq 'val'", true, nil}, // paging with mongo id
		{"gtin eq '123'", true, nil},                                       // key name with operator substring
		{"name eq 'val')", false, errors.New("")},                          // bad parentheses
		{"(name eq )", false, errors.New("")},                              // missing value in equals operator
		{"(name eq hello) and (name fakeop hello)", false, errors.New("")}, // bad operator
		{"epc_item_type ne 0 and name", false, errors.New("")},             // operators can't have a mix operators and literals
		{"epc_item_type ne 0 and 0", false, errors.New("")},                // // operators can't have a mix operators and literals
		{"name and epc_item_type ne 0", false, errors.New("")},             // // operators can't have a mix operators and literals
		{"name eqs epc_item_type", false, errors.New("")},                  // typo operator
		{"contains(and, epc_item_type)", false, errors.New("")},            // operator in function
		{"", false, errors.New("")},                                        // empty string test
	}

	for _, expectedVal := range filterTests {
		var testURLString = fmt.Sprintf("http://localhost/test?$filter=%s", expectedVal.input)
		testURL, err := url.Parse(testURLString)
		if err != nil {
			t.Fatalf("Failed to parse test url")
			return
		}

		parsedResult, err := ParseURLValues(testURL.Query())

		// an error happened, is it expected?
		if (err != nil) != (expectedVal.Err != nil) {
			t.Errorf("Expected error mismatch")
		}
		// else there is no error, this is ok
		if (parsedResult[Filter] == nil) == expectedVal.expected {
			t.Errorf("Expected: filter existance to be %t but it was %t", expectedVal.expected, !expectedVal.expected)
		}
	}

	// printTree(result[Filter].(*ParseNode), 0)
}

// func printTree(n *ParseNode, level int) {
// 	indent := ""
// 	for i := 0; i < level; i++ {
// 		indent += "  "
// 	}
// 	fmt.Printf("%s %-10s\n", indent, n.Token.Value)
// 	for _, v := range n.Children {
// 		printTree(v, level+1)
// 	}
// }
