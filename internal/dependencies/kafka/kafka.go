/*
 * Copyright (c) 2023 Asim Ihsan.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package kafka

import (
	"embed"
	"fmt"
	"github.com/cbroglie/mustache"
	"strings"
)

//go:embed *.mustache
var fs embed.FS

func GenerateDockerComposeFile(kafkaPort int) (string, error) {
	// valid port is between 1 and 65535
	if kafkaPort < 1 || kafkaPort > 65535 {
		return "", fmt.Errorf("invalid port number: %d", kafkaPort)
	}

	// Read the embedded template file
	templateFile, err := fs.ReadFile("docker-compose-template.mustache")
	if err != nil {
		return "", err
	}

	// Create a new template and parse the content
	tmpl, err := mustache.ParseStringRaw(string(templateFile), true)
	if err != nil {
		return "", err
	}

	parameters := make(map[string]string)
	parameters["kafka_port"] = fmt.Sprintf("%d", kafkaPort)

	// Render the template
	var buf strings.Builder
	if err = tmpl.FRender(&buf, parameters); err != nil {
		return "", err
	}

	// Return the generated content
	return buf.String(), nil
}
