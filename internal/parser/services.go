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
	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	parser "github.com/asimihsan/virtual-cluster/generated/services/grammar"
	"github.com/asimihsan/virtual-cluster/internal/utils"
	"strconv"
)

type ServicesAST struct {
	Services []ServiceAST
}

type ServiceAST struct {
	Name       string
	Repository *string
	Branch     *string
	Tag        *string
	Commit     *string
	Directory  *string
}

type servicesListener struct {
	parser.BaseServicesListener
	ast   *ServicesAST
	error error
}

func (l *servicesListener) EnterServicesConfig(ctx *parser.ServicesConfigContext) {
	l.ast = &ServicesAST{}
}

func (l *servicesListener) EnterServiceEntry(ctx *parser.ServiceEntryContext) {
	service := ServiceAST{}
	l.ast.Services = append(l.ast.Services, service)
}

func (l *servicesListener) EnterServiceName(ctx *parser.ServiceNameContext) {
	serviceName := ctx.IDENTIFIER().GetText()
	l.ast.Services[len(l.ast.Services)-1].Name = serviceName
}

func (l *servicesListener) EnterRepository(ctx *parser.RepositoryContext) {
	value := utils.HandleStringLiteral(ctx.STRING_LITERAL().GetText())
	l.ast.Services[len(l.ast.Services)-1].Repository = &value
}

func (l *servicesListener) EnterBranch(ctx *parser.BranchContext) {
	value := utils.HandleStringLiteral(ctx.STRING_LITERAL().GetText())
	l.ast.Services[len(l.ast.Services)-1].Branch = &value
}

func (l *servicesListener) EnterTag(ctx *parser.TagContext) {
	value := utils.HandleStringLiteral(ctx.STRING_LITERAL().GetText())
	l.ast.Services[len(l.ast.Services)-1].Tag = &value
}

func (l *servicesListener) EnterCommit(ctx *parser.CommitContext) {
	value := utils.HandleStringLiteral(ctx.STRING_LITERAL().GetText())
	l.ast.Services[len(l.ast.Services)-1].Commit = &value
}

func (l *servicesListener) EnterDirectory(ctx *parser.DirectoryContext) {
	value := utils.HandleStringLiteral(ctx.STRING_LITERAL().GetText())
	l.ast.Services[len(l.ast.Services)-1].Directory = &value
}

type errorListenerType struct {
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

func (r *errorListenerType) SyntaxError(
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

func ParseServices(input string) (*ServicesAST, error) {
	// Create the input stream and initialize the lexer and parser
	errorListener := &errorListenerType{}

	inputStream := antlr.NewInputStream(input)
	lexer := parser.NewServicesLexer(inputStream)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(errorListener)

	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	servicesParser := parser.NewServicesParser(tokenStream)
	servicesParser.RemoveErrorListeners()
	servicesParser.AddErrorListener(errorListener)

	// Create the listener and walk the tree starting at ServicesConfig
	listener := &servicesListener{}
	antlr.NewParseTreeWalker().Walk(listener, servicesParser.ServicesConfig())

	if listener.error != nil {
		return nil, listener.error
	}
	if errorListener.error {
		return nil, SyntaxErrors{Errors: errorListener.errors}
	}

	return listener.ast, nil
}
