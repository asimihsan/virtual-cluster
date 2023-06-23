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
	parser "github.com/asimihsan/virtual-cluster/generated/vcluster"
	"github.com/asimihsan/virtual-cluster/internal/utils"
	"strconv"
)

type VClusterAST struct {
	Services []VClusterDefinitionAST
}

type VClusterDefinitionAST struct {
	Name         string
	Dependencies []VClusterDependency
	HealthChecks VClusterHealthCheck
	StartupSeq   []VClusterStartupSequence
}

type VClusterDependency struct {
	Name string
}

type VClusterHealthCheck struct {
	Endpoint string
}

type VClusterStartupSequence struct {
	Command string
}

type vclusterListener struct {
	parser.BaseVClusterListener
	ast   *VClusterAST
	error error
}

func (l *vclusterListener) EnterVclusterConfig(ctx *parser.VclusterConfigContext) {
	l.ast = &VClusterAST{}
}

func (l *vclusterListener) EnterServiceConfig(ctx *parser.ServiceConfigContext) {
	serviceConfig := VClusterDefinitionAST{}
	l.ast.Services = append(l.ast.Services, serviceConfig)
}

func (l *vclusterListener) EnterServiceName(ctx *parser.ServiceNameContext) {
	serviceName := ctx.IDENTIFIER().GetText()
	l.ast.Services[len(l.ast.Services)-1].Name = serviceName
}

func (l *vclusterListener) EnterDependencyConfigItem(ctx *parser.DependencyConfigItemContext) {
	dependency := VClusterDependency{}
	l.ast.Services[len(l.ast.Services)-1].Dependencies = append(l.ast.Services[len(l.ast.Services)-1].Dependencies, dependency)
}

func (l *vclusterListener) EnterHealthCheckConfigItem(ctx *parser.HealthCheckConfigItemContext) {
	healthCheck := VClusterHealthCheck{}
	l.ast.Services[len(l.ast.Services)-1].HealthChecks = healthCheck
}

func (l *vclusterListener) EnterStartupSequenceConfigItem(ctx *parser.StartupSequenceConfigItemContext) {
	startupSequence := VClusterStartupSequence{}
	l.ast.Services[len(l.ast.Services)-1].StartupSeq = append(l.ast.Services[len(l.ast.Services)-1].StartupSeq, startupSequence)
}

func (l *vclusterListener) EnterDependencyName(ctx *parser.DependencyNameContext) {
	dependencyName := ctx.IDENTIFIER().GetText()
	l.ast.Services[len(l.ast.Services)-1].Dependencies[len(l.ast.Services[len(l.ast.Services)-1].Dependencies)-1].Name = dependencyName
}

func (l *vclusterListener) EnterEndpointHealthCheck(ctx *parser.EndpointHealthCheckContext) {
	stringLiteral := ctx.STRING_LITERAL()
	if stringLiteral == nil {
		return
	}
	value := utils.HandleStringLiteral(stringLiteral.GetText())
	l.ast.Services[len(l.ast.Services)-1].HealthChecks.Endpoint = value
}

func (l *vclusterListener) EnterCommandStartupSequence(ctx *parser.CommandStartupSequenceContext) {
	stringLiteral := ctx.STRING_LITERAL()
	if stringLiteral == nil {
		return
	}
	value := utils.HandleStringLiteral(stringLiteral.GetText())
	l.ast.Services[len(l.ast.Services)-1].StartupSeq[len(l.ast.Services[len(l.ast.Services)-1].StartupSeq)-1].Command = value
}

type vclusterErrorListenerType struct {
	*antlr.DefaultErrorListener
	errors []string
	error  bool
}

func (r *vclusterErrorListenerType) SyntaxError(
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

type VClusterSyntaxErrors struct {
	Errors []string
}

func ParseVCluster(input string) (*VClusterAST, error) {
	// Create the input stream and initialize the lexer and parser
	errorListener := &vclusterErrorListenerType{}

	inputStream := antlr.NewInputStream(input)
	lexer := parser.NewVClusterLexer(inputStream)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(errorListener)

	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	vClusterParser := parser.NewVClusterParser(tokenStream)
	vClusterParser.RemoveErrorListeners()
	vClusterParser.AddErrorListener(errorListener)

	// Create the listener and walk the tree starting at ServicesConfig
	listener := &vclusterListener{}

	antlr.NewParseTreeWalker().Walk(listener, vClusterParser.VclusterConfig())

	if listener.error != nil {
		return nil, listener.error
	}
	if errorListener.error {
		return nil, SyntaxErrors{Errors: errorListener.errors}
	}

	return listener.ast, nil
}
