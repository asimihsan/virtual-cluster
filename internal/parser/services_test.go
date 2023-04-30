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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseServices(t *testing.T) {
	input := `
// C++ style comment
/* C style comment */
# Python style comment

service myService {
  repository: "https://github.com/example/myService.git";
  branch: "main";
  directory: "/src";
}

service anotherService {
  repository: "https://github.com/example/anotherService.git";
  tag: "v1.0.0";
  directory: "/app";
}

service thirdService {
  repository: "https://github.com/example/thirdService.git";
  commit: "3a5b2c1d";
  directory: "/";
}
`
	ast, err := ParseServices(input)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(ast.Services))

	assert.Equal(t, "myService", ast.Services[0].Name)
	assert.Equal(t, "https://github.com/example/myService.git", *ast.Services[0].Repository)
	assert.Equal(t, "main", *ast.Services[0].Branch)
	assert.Nil(t, ast.Services[0].Tag)
	assert.Nil(t, ast.Services[0].Commit)
	assert.Equal(t, "/src", *ast.Services[0].Directory)

	assert.Equal(t, "anotherService", ast.Services[1].Name)
	assert.Equal(t, "https://github.com/example/anotherService.git", *ast.Services[1].Repository)
	assert.Nil(t, nil, ast.Services[1].Branch)
	assert.Equal(t, "v1.0.0", *ast.Services[1].Tag)
	assert.Nil(t, nil, ast.Services[1].Commit)
	assert.Equal(t, "/app", *ast.Services[1].Directory)

	assert.Equal(t, "thirdService", ast.Services[2].Name)
	assert.Equal(t, "https://github.com/example/thirdService.git", *ast.Services[2].Repository)
	assert.Nil(t, ast.Services[2].Branch)
	assert.Nil(t, ast.Services[2].Tag)
	assert.Equal(t, "3a5b2c1d", *ast.Services[2].Commit)
	assert.Equal(t, "/", *ast.Services[2].Directory)
}

func TestParseServices_Error(t *testing.T) {
	input := `
service myService {
  repository: "https://github.com/example/myService.git";
  branch: main;
  directory: "/src";
}

service anotherService {
  repository: "https://github.com/example/anotherService.git";
  tag: v1.0.0;
  directory: "/app";
}

service thirdService {
  repository: "https://github.com/example/thirdService.git";
  commit: 3a5b2c1d;
  directory: "/";  // missing closing brace
`
	ast, err := ParseServices(input)
	assert.Error(t, err)
	assert.Nil(t, ast)
}
