/* Apache v2 license
*  Copyright (C) <2019> Intel Corporation
*
*  SPDX-License-Identifier: Apache-2.0
 */

package validatefield

type validField struct {
	characters map[string]bool
}

func New(characters string) *validField {

	var obj = validField{make(map[string]bool, len(characters))}

	// build map
	for _, val := range characters {
		obj.characters[string(val)] = true
	}

	return &obj
}

func (v *validField) ValidateField(value string) bool {
	return v.characters[value]
}
