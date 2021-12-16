/* Apache v2 license
*  Copyright (C) <2019> Intel Corporation
*
*  SPDX-License-Identifier: Apache-2.0
 */

package parser

// Token constants
const (
	filterTokenOpenParen int = iota
	filterTokenCloseParen
	filterTokenWhitespace
	filterTokenComma
	filterTokenLogical
	filterTokenFunc
	filterTokenFloat
	filterTokenInteger
	filterTokenString
	filterTokenDate
	filterTokenTime
	filterTokenDateTime
	filterTokenBoolean
	filterTokenLiteral
)

// GlobalFilterTokenizer the global filter tokenizer
var globalFilterTokenizer = filterTokenizer()

// GlobalFilterParser the global filter parser
var globalFilterParser = filterParser()

// ParseFilterString Converts an input string from the $filter part of the URL into a parse
// tree that can be used by providers to create a response.
func parseFilterString(filter string) (*ParseNode, error) {
	tokens, err := globalFilterTokenizer.tokenize(filter)
	if err != nil {
		return nil, err
	}
	// TODO: can we do this in one fell swoop?
	postfix, err := globalFilterParser.infixToPostfix(tokens)
	if err != nil {
		return nil, err
	}
	tree, err := globalFilterParser.postfixToTree(postfix)
	if err != nil {
		return nil, err
	}

	return tree, nil
}

// FilterTokenizer Creates a tokenizer capable of tokenizing filter statements
func filterTokenizer() *Tokenizer {
	t := Tokenizer{}
	t.add("^\\(", filterTokenOpenParen)
	t.add("^\\)", filterTokenCloseParen)
	t.add("^,", filterTokenComma)
	t.add("^(eq|ne|gt|ge|lt|le|and|or) ", filterTokenLogical)
	t.add("^(contains|endswith|startswith)", filterTokenFunc)
	t.add("^-?[0-9]+\\.[0-9]+", filterTokenFloat)
	t.add("^-?[0-9]+$", filterTokenInteger)
	t.add("^(?i:true|false)", filterTokenBoolean)
	t.add("^'(''|[^'])*'", filterTokenString)
	t.add("^[0-9]{4,4}-[0-9]{2,2}-[0-9]{2,2}T[0-9]{2,2}:[0-9]{2,2}(:[0-9]{2,2}(.[0-9]+)?)?(Z|[+-][0-9]{2,2}:[0-9]{2,2})", filterTokenDateTime)
	t.add("^-?[0-9]{4,4}-[0-9]{2,2}-[0-9]{2,2}", filterTokenDate)
	//t.add("^[0-9]{2,2}:[0-9]{2,2}(:[0-9]{2,2}(.[0-9]+)?)?", filterTokenTime)
	t.add("^[a-zA-Z][a-zA-Z0-9_.]*", filterTokenLiteral)
	t.add("^_id", filterTokenLiteral)
	t.ignore("^ ", filterTokenWhitespace)

	return &t
}

// FilterParser creates the definitions for operators and functions
func filterParser() *Parser {
	parser := emptyParser()
	parser.defineOperator("gt", 2, opAssociationLeft, 4)
	parser.defineOperator("ge", 2, opAssociationLeft, 4)
	parser.defineOperator("lt", 2, opAssociationLeft, 4)
	parser.defineOperator("le", 2, opAssociationLeft, 4)
	parser.defineOperator("eq", 2, opAssociationLeft, 3)
	parser.defineOperator("ne", 2, opAssociationLeft, 3)
	parser.defineOperator("and", 2, opAssociationLeft, 2)
	parser.defineOperator("or", 2, opAssociationLeft, 1)
	parser.defineFunction("contains", 2)
	parser.defineFunction("endswith", 2)
	parser.defineFunction("startswith", 2)

	return parser
}
