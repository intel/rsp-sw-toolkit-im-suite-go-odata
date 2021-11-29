/* Apache v2 license
*  Copyright (C) <2019> Intel Corporation
*
*  SPDX-License-Identifier: Apache-2.0
 */

package example

import (
	"fmt"
	"net/url"

	"github.com/globalsign/mgo"
	odata "github.com/intel/rsp-sw-toolkit-im-suite-go-odata/mongo"
)

func example() {

	var dbhost = "mongodb://localhost:27017/testdb"

	mainSession, err := mgo.Dial(dbhost)
	if err != nil {
		fmt.Errorf("Unable to connect to mongo server on %s", dbhost)
	}

	defer mainSession.Close()

	testURL, err := url.Parse("http://127.0.0.1/test?$top=10&$select=name,age&$orderby=time asc,name desc,age")
	if err != nil {
		fmt.Errorf("failed to parse test url")
	}

	var object []interface{}
	collection := mainSession.DB("testdb").C("collectionName")

	if err := odata.ODataQuery("", testURL.Query(), &object, collection); err != nil {
		fmt.Errorf("Error: %s", err.Error())
	}
}
