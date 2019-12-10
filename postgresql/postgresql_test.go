/* Apache v2 license
*  Copyright (C) <2019> Intel Corporation
*
*  SPDX-License-Identifier: Apache-2.0
 */

package postgresql

import (
	"database/sql"
	"fmt"
	"net/url"
	"testing"

	_ "github.com/lib/pq" // postgreSQL driver
	"github.com/pkg/errors"
)

// For testing, run PostgreSQL as a docker container
// docker run -p 5432:5432 postgres:11-alpine
// current SQL mock testing libraries do not support jsonb

const (
	host     = "postgres"
	port     = 5432
	user     = "postgres"
	password = ""
	dbname   = "postgres"
)

func TestFilter(t *testing.T) {

	db := dbSetup()

	var filterTests = []struct {
		input    string
		expected bool
		Err      error
	}{
		{"((epc_item_type gt 0) and (event ne 'departed') and (required eq true) and (count lt 0.1) or " +
			"(SKU eq '123') and contains(epc_time, '0') and startswith(epc_code, '456') or " +
			"endswith(upc_code, '789'))", true, nil}, // large valid filter case
		{"name eq 'test'", true, nil},                                      // key name with operator
		{"age gt 10", true, nil},                                           // key name with operator
		{"age ge 10", true, nil},                                           // key name with operator
		{"age lt 10", true, nil},                                           // key name with operator
		{"age le 10", true, nil},                                           // key name with operator
		{"(name eq 'test') and (age gt 10) or (age le 10)", true, nil},     // key name with operator
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

	for _, expectedVal := range filterTests {
		var testURLString = fmt.Sprintf("http://localhost/test?$filter=%s", expectedVal.input)
		testURL, err := url.Parse(testURLString)
		if err != nil {
			t.Fatalf("Failed to parse test url")
			return
		}

		_, errorQuery := ODataSQLQuery(testURL.Query(), "test", "data", db)

		if (errorQuery != nil) != (expectedVal.Err != nil) {
			t.Errorf("Expected error mismatch. Error : %s", errorQuery)
		}

	}

}

func TestODataQuery(t *testing.T) {

	db := dbSetup()

	testURL, err := url.Parse("http://localhost/test?$top=10&$skip=10&$select=name,lastname,age&$orderby=time asc,name desc,age")
	if err != nil {
		t.Error("failed to parse test url")
	}

	_, errorQuery := ODataSQLQuery(testURL.Query(), "test", "data", db)

	if errorQuery != nil {
		t.Error(errorQuery)
	}

}

func TestCount(t *testing.T) {

	db := dbSetup()

	_, errorQuery := ODataCount(db, "test")

	if errorQuery != nil {
		t.Error(errorQuery)
	}

}

func dbSetup() *sql.DB {

	const schema = `
			CREATE TABLE IF NOT EXISTS test (	
				id int,
				data JSONB	
			);
	`
	// Connect to PostgreSQL
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	// Create table
	db.Exec(schema)

	return db
}
