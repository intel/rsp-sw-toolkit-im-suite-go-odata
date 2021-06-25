/* Apache v2 license
*  Copyright (C) <2019> Intel Corporation
*
*  SPDX-License-Identifier: Apache-2.0
 */

package parser

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Operator constants
const (
	opAssociationLeft int = iota
	opAssociationRight
)

// Tokenizer structure
type Tokenizer struct {
	TokenMatchers  []*TokenMatcher
	IgnoreMatchers []*TokenMatcher
}

// TokenMatcher token matcher structure
type TokenMatcher struct {
	Pattern string
	Re      *regexp.Regexp
	Token   int
}

// Token token structure
type Token struct {
	stringValue string
	Value       interface{}
	Type        int
}

// Add adds token to the tokenizer
func (t *Tokenizer) add(pattern string, token int) {
	rxp := regexp.MustCompile(pattern)
	matcher := &TokenMatcher{pattern, rxp, token}
	t.TokenMatchers = append(t.TokenMatchers, matcher)
}

// Ignore adds ignore case to the tokenizer
func (t *Tokenizer) ignore(pattern string, token int) {
	rxp := regexp.MustCompile(pattern)
	matcher := &TokenMatcher{pattern, rxp, token}
	t.IgnoreMatchers = append(t.IgnoreMatchers, matcher)
}

// TokenizeBytes tokenizes the bytes
func (t *Tokenizer) tokenizeBytes(target []byte) ([]*Token, error) {
	result := make([]*Token, 0)
	match := true // false when no match is found
	for len(target) > 0 && match {
		match = false
		for _, m := range t.TokenMatchers {
			token := m.Re.Find(target)
			if len(token) > 0 {
				convValue, _ := convertValue(token, m.Token)
				parsed := Token{stringValue: strings.TrimSpace(string(token)), Value: convValue, Type: m.Token}
				result = append(result, &parsed)
				target = target[len(token):] // remove the token from the input
				match = true
				break
			}
		}
		for _, m := range t.IgnoreMatchers {
			token := m.Re.Find(target)
			if len(token) > 0 {
				match = true
				target = target[len(token):] // remove the token from the input
				break
			}
		}
	}

	if len(target) > 0 && !match {
		return result, errors.New("No matching token for " + string(target))
	}

	return result, nil
}

func convertValue(token []byte, tokenType int) (interface{}, error) {
	switch tokenType {
	case filterTokenInteger:
		return strconv.Atoi(string(token))
	case filterTokenBoolean:
		return strconv.ParseBool(string(token))
	case filterTokenFloat:
		return strconv.ParseFloat(string(token), 10)
	case filterTokenLiteral, filterTokenString:
		return strings.TrimSpace(string(token)), nil
	case filterTokenDateTime, filterTokenDate, filterTokenTime:
		return time.Parse("2006-01-02", string(token))
	default:
		return strings.TrimSpace(string(token)), nil
	}
}

// Tokenize tokenize string by converting it to bytes and passing the array to the tokeizeBytes function
func (t *Tokenizer) tokenize(target string) ([]*Token, error) {
	return t.tokenizeBytes([]byte(target))
}

// Parser parser structure
type Parser struct {
	// Map from string inputs to operator types
	Operators map[string]*Operator
	// Map from string inputs to function types
	Functions map[string]*Function
}

// Operator operator structure
type Operator struct {
	Token string
	// Whether the operator is left/right/or not associative
	Association int
	// The number of operands this operator operates on
	Operands int
	// Rank of precedence
	Precedence int
}

// Function function structure
type Function struct {
	Token string
	// The number of parameters this function accepts
	Params int
}

// ParseNode parseNode structure
type ParseNode struct {
	Token    *Token
	Parent   *ParseNode
	Children []*ParseNode
}

// EmptyParser create empty parser
func emptyParser() *Parser {
	return &Parser{make(map[string]*Operator), make(map[string]*Function)}
}

// DefineOperator Adds an operator to the language. Provide the token, a precedence, and
// whether the operator is left, right, or not associative.
func (p *Parser) defineOperator(token string, operands, assoc, precedence int) {
	p.Operators[token] = &Operator{token, assoc, operands, precedence}
}

// DefineFunction Adds a function to the language
func (p *Parser) defineFunction(token string, params int) {
	p.Functions[token] = &Function{token, params}
}

// InfixToPostfix Parses the input string of tokens using the given definitions of operators
// and functions. (Everything else is assumed to be a literal.) Uses the
// Shunting-Yard algorithm.
//nolint :gocyclo
func (p *Parser) infixToPostfix(tokens []*Token) (*tokenQueue, error) {
	queue := tokenQueue{}
	stack := tokenStack{}
	wasLiteral := false // We use this bool to see if the last token was a literal

	for len(tokens) > 0 {
		token := tokens[0]
		tokens = tokens[1:]

		if _, ok := p.Functions[token.stringValue]; ok {
			// push functions onto the stack
			stack.push(token)
			wasLiteral = false
		} else if token.stringValue == "," {
			// function parameter separator, pop off stack until we see a "("
			for stack.peek().stringValue != "(" || stack.empty() {
				queue.enqueue(stack.pop())
			}
			// there was an error parsing
			if stack.empty() {
				return nil, errors.New("Parse error")
			}
			wasLiteral = false
		} else if o1, ok := p.Operators[token.stringValue]; ok {
			// push operators onto stack according to precedence
			if !stack.empty() {
				for o2, ok := p.Operators[stack.peek().stringValue]; ok &&
					(o1.Association == opAssociationLeft && o1.Precedence <= o2.Precedence) ||
					(o1.Association == opAssociationRight && o1.Precedence < o2.Precedence); {
					queue.enqueue(stack.pop())

					if stack.empty() {
						break
					}
					o2, ok = p.Operators[stack.peek().stringValue]
				}
			}
			stack.push(token)
			wasLiteral = false
		} else if token.stringValue == "(" {
			// push open parens onto the stack
			stack.push(token)
			wasLiteral = false
		} else if token.stringValue == ")" {
			// if we find a close paren, pop things off the stack
			for !stack.empty() && stack.peek().stringValue != "(" {
				queue.enqueue(stack.pop())
			}
			// there was an error parsing
			if stack.empty() {
				return nil, errors.New("parse error: mismatched parenthesis")
			}
			// pop off open paren
			stack.pop()
			// if next token is a function, move it to the queue
			if !stack.empty() {
				if _, ok := p.Functions[stack.peek().stringValue]; ok {
					queue.enqueue(stack.pop())
				}
			}
			wasLiteral = false
		} else {
			// if the last token was a literal it means we are trying to push 2 literals into the queue back to back
			// This will cause issues in the tree parsing. This is a rules violation and will throw an error
			if wasLiteral {
				return nil, errors.New("parse error: two literals found in a row")
			}
			// Token is a literal -- put it in the queue and set the bool to true
			queue.enqueue(token)
			wasLiteral = true
		}
	}

	// pop off the remaining operators onto the queue
	for !stack.empty() {
		if stack.peek().stringValue == "(" || stack.peek().stringValue == ")" {
			return nil, errors.New("parse error: mismatched parenthesis")
		}
		queue.enqueue(stack.pop())
	}

	return &queue, nil
}

// PostfixToTree Converts a Postfix token queue to a parse tree
//nolint :gocyclo
func (p *Parser) postfixToTree(queue *tokenQueue) (*ParseNode, error) {
	stack := &nodeStack{}
	currNode := &ParseNode{}

	t := queue.Head
	for t != nil {
		t = t.Next
	}

	for !queue.empty() {
		// push the token onto the stack as a tree node
		currNode = &ParseNode{queue.dequeue(), nil, make([]*ParseNode, 0)}
		stack.push(currNode)

		if _, ok := p.Functions[stack.peek().Token.stringValue]; ok {
			// if the top of the stack is a function
			node, err := stack.pop()
			if err != nil {
				return nil, err
			}
			f := p.Functions[node.Token.stringValue]

			// pop off function parameters
			for i := 0; i < f.Params; i++ {
				childNode, childErr := stack.pop()
				if childErr != nil {
					return nil, childErr
				}
				// prepend children so they get added in the right order
				node.Children = append([]*ParseNode{childNode}, node.Children...)
			}

			if !checkChildType(node.Children) {
				return nil, errors.New("Cannot have literal and function/operator mismatch")
			}
			stack.push(node)
		} else if _, ok := p.Operators[stack.peek().Token.stringValue]; ok {
			// if the top of the stack is an operator
			node, err := stack.pop()
			if err != nil {
				return nil, err
			}
			o := p.Operators[node.Token.stringValue]

			// pop off operands
			for i := 0; i < o.Operands; i++ {
				// prepend children so they get added in the right order
				childNode, childErr := stack.pop()
				if childErr != nil {
					return nil, childErr
				}
				node.Children = append([]*ParseNode{childNode}, node.Children...)
			}
			if !checkChildType(node.Children) {
				return nil, errors.New("Cannot have literal and function/operator mismatch")
			}
			stack.push(node)
		}
	}

	return currNode, nil
}

// checkChildType Checks to make sure children types are compatible
//nolint :gocyclo
func checkChildType(child []*ParseNode) bool {
	// Make sure we have 2 children in the tree
	if len(child) != 2 {
		return false
	}

	// Make sure that the token struct exists
	for _, c := range child {
		if c.Token == nil {
			return false
		}
	}

	// Chrildren with the same type are compatible
	if child[0].Token.Type == child[1].Token.Type {
		return true
	}
	// If the first child is not an operator and function and
	// the second child is we have an invalid combination
	if (child[0].Token.Type != filterTokenLogical &&
		child[0].Token.Type != filterTokenFunc) &&
		(child[1].Token.Type == filterTokenLogical ||
			child[1].Token.Type == filterTokenFunc) {
		return false
	}
	// If the second child is not an operator and function and
	// the first child is we have an invalid combination
	if (child[0].Token.Type == filterTokenLogical ||
		child[0].Token.Type == filterTokenFunc) &&
		(child[1].Token.Type != filterTokenLogical &&
			child[1].Token.Type != filterTokenFunc) {
		return false
	}
	return true
}

type tokenStack struct {
	Head *tokenStackNode
	Size int
}

type tokenStackNode struct {
	Token *Token
	Prev  *tokenStackNode
}

func (s *tokenStack) push(t *Token) {
	node := tokenStackNode{t, s.Head}
	s.Head = &node
	s.Size++
}

func (s *tokenStack) pop() *Token {
	node := s.Head
	s.Head = node.Prev
	s.Size--
	return node.Token
}

func (s *tokenStack) peek() *Token {
	return s.Head.Token
}

func (s *tokenStack) empty() bool {
	return s.Head == nil
}

type tokenQueue struct {
	Head *tokenQueueNode
	Tail *tokenQueueNode
}

type tokenQueueNode struct {
	Token *Token
	Prev  *tokenQueueNode
	Next  *tokenQueueNode
}

func (q *tokenQueue) enqueue(t *Token) {
	node := tokenQueueNode{t, q.Tail, nil}
	//fmt.Println(t.Value)

	if q.Tail == nil {
		q.Head = &node
	} else {
		q.Tail.Next = &node
	}

	q.Tail = &node
}

func (q *tokenQueue) dequeue() *Token {
	node := q.Head
	if node.Next != nil {
		node.Next.Prev = nil
	}
	q.Head = node.Next
	if q.Head == nil {
		q.Tail = nil
	}
	return node.Token
}

func (q *tokenQueue) empty() bool {
	return q.Head == nil && q.Tail == nil
}

type nodeStack struct {
	Head *nodeStackNode
}

type nodeStackNode struct {
	ParseNode *ParseNode
	Prev      *nodeStackNode
}

func (s *nodeStack) push(n *ParseNode) {
	node := nodeStackNode{n, s.Head}
	s.Head = &node
}

func (s *nodeStack) pop() (*ParseNode, error) {
	node := s.Head
	if node == nil {
		return nil, errors.New("Child node not available")
	}
	s.Head = node.Prev
	return node.ParseNode, nil
}

func (s *nodeStack) peek() *ParseNode {
	return s.Head.ParseNode
}
