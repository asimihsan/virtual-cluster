/*
 * Copyright (c) 2023 Asim Ihsan.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/asimihsan/virtual-cluster/internal/utils"
)

func TestStatUpward(t *testing.T) {
	// Test that a file in the current directory can be found.
	stat, path, err := utils.StatUpward("os.go", false)
	if err != nil {
		t.Errorf("StatUpward returned an error: %v", err)
	}
	if !stat.Mode().IsRegular() {
		t.Errorf("StatUpward returned a non-regular file: %v", stat)
	}
	if filepath.Base(path) != "os.go" {
		t.Errorf("StatUpward returned an incorrect path: %v", path)
	}

	// Test that a file in a parent directory can be found.
	stat, path, err = utils.StatUpward("README.md", false)
	if err != nil {
		t.Errorf("StatUpward returned an error: %v", err)
	}
	if !stat.Mode().IsRegular() {
		t.Errorf("StatUpward returned a non-regular file: %v", stat)
	}
	if filepath.Base(path) != "README.md" {
		t.Errorf("StatUpward returned an incorrect path: %v", path)
	}

	// Test that a non-existent file returns an error.
	stat, path, err = utils.StatUpward("non-existent-file", false)
	if err == nil {
		t.Errorf("StatUpward did not return an error for a non-existent file")
	}
	if stat != nil {
		t.Errorf("StatUpward returned a non-nil stat for a non-existent file: %v", stat)
	}
	if path != "" {
		t.Errorf("StatUpward returned a non-empty path for a non-existent file: %v", path)
	}

	// Test that an absolute path can be found.
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("Failed to get working directory: %v", err)
	}
	stat, path, err = utils.StatUpward(wd, false)
	if err != nil {
		t.Errorf("StatUpward returned an error: %v", err)
	}
	if !stat.Mode().IsDir() {
		t.Errorf("StatUpward returned a non-directory file: %v", stat)
	}
	if path != wd {
		t.Errorf("StatUpward returned an incorrect path: %v", path)
	}
}
