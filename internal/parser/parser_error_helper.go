/*
 * Copyright (c) 2023 Asim Ihsan.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package parser

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"strconv"
	"strings"
)

type ErrorListenerType struct {
	*antlr.DefaultErrorListener
	errors []string
	error  bool
}

type SyntaxErrors struct {
	Errors []string
}

func (e SyntaxErrors) Error() string {
	return fmt.Sprintf("syntax errors: %v", e.Errors)
}

func (r *ErrorListenerType) SyntaxError(
	recognizer antlr.Recognizer,
	offendingSymbol interface{},
	line, column int,
	msg string,
	e antlr.RecognitionException,
) {
	errorString := fmt.Sprintf("line " + strconv.Itoa(line) + ":" + strconv.Itoa(column) + " " + msg)
	r.errors = append(r.errors, errorString)
	r.error = true
}

type ParserErrorHelper struct {
	lines []string
}

func NewParserErrorHelper(input string) *ParserErrorHelper {
	lines := strings.Split(input, "\n")
	return &ParserErrorHelper{
		lines: lines,
	}
}

// FriendlyError is a struct that contains terse error like line, col, short message, but also
// a longer message that is more helpful to the user.
//
// e.g.
//
// error[E0308]: token recognition error
//  --> line 23 column 9
//    |
// 22 | /     if exit > cycles {
// 23 | |         cycles
//    | |         ^^^^^^ expected `()`, found `u8`
// 24 | |     }
//    | |_____- expected this to be `()`
//    |
// help: you might have meant to return this value
//type FriendlyError struct {
//	Line   int
//	Column int
//	ShortMsg    string
//	LongMsg     string
//}
//
//func (e *ParserErrorHelper) GetFriendlyError(
//	node antlr.ErrorNode,
//) FriendlyError {
//	// Get the line and column number.
//	line := node.GetSymbol().GetLine()
//	column := node.GetSymbol().GetColumn()
//
//	// Get the short message.
//	shortMsg := node.GetMessage()
//}
