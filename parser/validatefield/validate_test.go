/* Apache v2 license
*  Copyright (C) <2019> Intel Corporation
*
*  SPDX-License-Identifier: Apache-2.0
 */

package validatefield

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {

	input := "!@#"

	obj := New(input)

	// Expected result
	expectedMap := make(map[string]bool, 3)
	expectedMap["!"] = true
	expectedMap["@"] = true
	expectedMap["#"] = true

	var expectedObj = validField{characters: expectedMap}

	if !reflect.DeepEqual(obj, &expectedObj) {
		t.Error("validfield obj is not the same as expected object")
	}

}
