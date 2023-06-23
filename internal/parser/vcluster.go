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
	"strconv"

	"github.com/antlr4-go/antlr/v4"
	parser "github.com/asimihsan/virtual-cluster/generated/vcluster"
	"github.com/asimihsan/virtual-cluster/internal/utils"
)

type VClusterAST struct {
	Services     []VClusterServiceDefinitionAST
	Dependencies []VClusterDependencyDefinitionAST
}

type VClusterServiceDefinitionAST struct {
	Name         string
	Repository   *string
	Branch       *string
	Tag          *string
	Commit       *string
	Directory    *string
	HealthChecks HealthCheck
	Dependencies []VClusterDependency
	RunCommands  []string
}

type VClusterDependency struct {
	Name string
}

type HealthCheck struct {
	Endpoint string
}

type VClusterDependencyDefinitionAST struct {
	Name         string
	HealthChecks HealthCheck
	Dependencies []VClusterDependency
}

type vclusterListener struct {
	parser.BaseVClusterListener
	ast   *VClusterAST
	error error
}

func (l *vclusterListener) EnterVclusterConfig(ctx *parser.VclusterConfigContext) {
	l.ast = &VClusterAST{}
}

func (l *vclusterListener) EnterConfigEntry(ctx *parser.ConfigEntryContext) {}

func (l *vclusterListener) EnterServiceEntry(ctx *parser.ServiceEntryContext) {
	serviceConfig := VClusterServiceDefinitionAST{}
	l.ast.Services = append(l.ast.Services, serviceConfig)
}

func (l *vclusterListener) EnterDependencyEntry(ctx *parser.DependencyEntryContext) {
	dependencyConfig := VClusterDependencyDefinitionAST{}
	l.ast.Dependencies = append(l.ast.Dependencies, dependencyConfig)
}

func (l *vclusterListener) EnterServiceName(ctx *parser.ServiceNameContext) {
	serviceName := ctx.IDENTIFIER().GetText()
	l.ast.Services[len(l.ast.Services)-1].Name = serviceName
}

func (l *vclusterListener) EnterDependencyName(ctx *parser.DependencyNameContext) {
	dependencyName := ctx.IDENTIFIER().GetText()
	l.ast.Dependencies[len(l.ast.Dependencies)-1].Name = dependencyName
}

func (l *vclusterListener) EnterServiceConfigRepository(ctx *parser.ServiceConfigRepositoryContext) {
	repository := ctx.STRING_LITERAL()
	if repository == nil {
		return
	}
	value := utils.HandleStringLiteral(repository.GetText())
	l.ast.Services[len(l.ast.Services)-1].Repository = &value
}

func (l *vclusterListener) EnterServiceConfigBranch(ctx *parser.ServiceConfigBranchContext) {
	branch := ctx.STRING_LITERAL()
	if branch == nil {
		return
	}
	value := utils.HandleStringLiteral(branch.GetText())
	l.ast.Services[len(l.ast.Services)-1].Branch = &value
}

func (l *vclusterListener) EnterServiceConfigTag(ctx *parser.ServiceConfigTagContext) {
	tag := ctx.STRING_LITERAL()
	if tag == nil {
		return
	}
	value := utils.HandleStringLiteral(tag.GetText())
	l.ast.Services[len(l.ast.Services)-1].Tag = &value
}

func (l *vclusterListener) EnterServiceConfigCommit(ctx *parser.ServiceConfigCommitContext) {
	commit := ctx.STRING_LITERAL()
	if commit == nil {
		return
	}
	value := utils.HandleStringLiteral(commit.GetText())
	l.ast.Services[len(l.ast.Services)-1].Commit = &value
}

func (l *vclusterListener) EnterServiceConfigDirectory(ctx *parser.ServiceConfigDirectoryContext) {
	directory := ctx.STRING_LITERAL()
	if directory == nil {
		return
	}
	value := utils.HandleStringLiteral(directory.GetText())
	l.ast.Services[len(l.ast.Services)-1].Directory = &value
}

func (l *vclusterListener) EnterServiceConfigHealthCheck(ctx *parser.ServiceConfigHealthCheckContext) {
	healthCheck := HealthCheck{}
	l.ast.Services[len(l.ast.Services)-1].HealthChecks = healthCheck
}

func (l *vclusterListener) EnterServiceConfigDependency(ctx *parser.ServiceConfigDependencyContext) {}

func (l *vclusterListener) EnterDependencyConfigHealthCheck(ctx *parser.DependencyConfigHealthCheckContext) {
	healthCheck := HealthCheck{}
	l.ast.Dependencies[len(l.ast.Dependencies)-1].HealthChecks = healthCheck
}

func (l *vclusterListener) EnterDependencyConfigDependency(ctx *parser.DependencyConfigDependencyContext) {
	dependency := VClusterDependency{}
	l.ast.Dependencies[len(l.ast.Dependencies)-1].Dependencies = append(l.ast.Dependencies[len(l.ast.Dependencies)-1].Dependencies, dependency)
}

func (l *vclusterListener) EnterHealthCheckEndpoint(ctx *parser.HealthCheckEndpointContext) {
	endpoint := ctx.STRING_LITERAL()
	if endpoint == nil {
		return
	}
	value := utils.HandleStringLiteral(endpoint.GetText())

	// TODO this is incorrect because this could also be a dependency healthcheck, how to handle?
	l.ast.Services[len(l.ast.Services)-1].HealthChecks.Endpoint = value
}

func (l *vclusterListener) EnterServiceConfigRunCommands(ctx *parser.ServiceConfigRunCommandsContext) {
	runCommands := ctx.AllSTRING_LITERAL()
	if runCommands == nil {
		return
	}
	for _, runCommand := range runCommands {
		value := utils.HandleStringLiteral(runCommand.GetText())
		l.ast.Services[len(l.ast.Services)-1].RunCommands = append(l.ast.Services[len(l.ast.Services)-1].RunCommands, value)
	}
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
