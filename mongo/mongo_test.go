/* Apache v2 license
*  Copyright (C) <2019> Intel Corporation
*
*  SPDX-License-Identifier: Apache-2.0
 */

package mongo

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/pkg/errors"

	"github.com/globalsign/mgo"
)

var dbhost = "mongodb://localhost:27017/test"

func TestODataQuery(t *testing.T) {

	mainSession, err := mgo.Dial(dbhost)
	if err != nil {
		t.Errorf("Unable to connect to mongo server on %s", dbhost)
	}

	testURL, err := url.Parse("http://localhost/test?$top=10&$select=name,age&$orderby=time asc,name desc,age")
	if err != nil {
		t.Error("failed to parse test url")
	}

	var object []interface{}
	collection := mainSession.DB("test").C("tests")

	if err := ODataQuery("", testURL.Query(), &object, collection); err != nil {
		t.Errorf("Error: %s", err.Error())
	}
}
func TestODataCount(t *testing.T) {

	mainSession, err := mgo.Dial(dbhost)
	if err != nil {
		t.Errorf("Unable to connect to mongo server on %s", dbhost)
	}

	collection := mainSession.DB("test").C("tests")

	if _, err := ODataCount(collection); err != nil {
		t.Errorf("Error: %s", err.Error())
	}
}

func TestODataFilter(t *testing.T) {

	mainSession, err := mgo.Dial(dbhost)
	if err != nil {
		t.Errorf("Unable to connect to mongo server on %s", dbhost)
	}

	testURL, err := url.Parse("http://localhost/test?$filter=notificationid eq 'TestInsert' and url eq 'test' or comment ne 'test'")
	if err != nil {
		t.Error("failed to parse test url")
	}

	var object []interface{}
	collection := mainSession.DB("test").C("tests")

	if err := ODataQuery("", testURL.Query(), &object, collection); err != nil {
		t.Errorf("Error: %s", err.Error())
	}
}

func TestODataFunctions(t *testing.T) {

	mainSession, err := mgo.Dial(dbhost)
	if err != nil {
		t.Errorf("Unable to connect to mongo server on %s", dbhost)
	}

	testURL, err := url.Parse("http://localhost/test?$filter=startswith(epc_code, '456') or endswith(upc_code, '789') or contains(SKU, '123')")
	if err != nil {
		t.Error("failed to parse test url")
	}

	var object []interface{}
	collection := mainSession.DB("test").C("tests")

	if err := ODataQuery("", testURL.Query(), &object, collection); err != nil {
		t.Errorf("Error: %s", err.Error())
	}

}

func TestODataFilterGreaterLessThan(t *testing.T) {

	mainSession, err := mgo.Dial(dbhost)
	if err != nil {
		t.Errorf("Unable to connect to mongo server on %s", dbhost)
	}

	testURL, err := url.Parse("http://localhost/test?$filter=price gt 20 or price lt 10 or price ge 30 or price le 50")
	if err != nil {
		t.Error("failed to parse test url")
	}

	var object []interface{}
	collection := mainSession.DB("test").C("tests")

	if err := ODataQuery("", testURL.Query(), &object, collection); err != nil {
		t.Errorf("Error: %s", err.Error())
	}
}

func TestODataWithFilter(t *testing.T) {
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
		{"0 eq epc_item_type", false, errors.New("")},                      // integer key name
		{"", false, errors.New("")},                                        // empty string test
	}

	mainSession, err := mgo.Dial(dbhost)
	if err != nil {
		t.Errorf("Unable to connect to mongo server on %s", dbhost)
	}
	var object []interface{}
	collection := mainSession.DB("test").C("tests")

	for _, expectedVal := range filterTests {
		var testURLString = fmt.Sprintf("http://localhost/test?$filter=%s", expectedVal.input)
		testURL, err := url.Parse(testURLString)
		if err != nil {
			t.Fatalf("Failed to parse test url")
			return
		}

		queryErr := ODataQuery("", testURL.Query(), &object, collection)

		// an error happened, is it expected?
		if (queryErr != nil) != (expectedVal.Err != nil) {
			t.Errorf("Expected error mismatch")
		}

	}
}
