/* Apache v2 license
*  Copyright (C) <2019> Intel Corporation
*
*  SPDX-License-Identifier: Apache-2.0
 */

package parser

import (
	"strconv"
	"strings"
)

func parseInt(value *string) (int, error) {
	result, err := strconv.Atoi(strings.TrimSpace(*value))
	if err != nil {
		return 0, err
	}
	return result, nil
}
