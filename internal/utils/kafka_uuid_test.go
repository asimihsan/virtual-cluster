/*
 * Copyright (c) 2023 Asim Ihsan.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package utils

import (
	"strings"
	"testing"
)

func TestRandomUUID(t *testing.T) {
	// Test that a random UUID is generated.
	uuid := randomUUID()
	if uuid.mostSignificantBits == 0 && uuid.leastSignificantBits == 0 {
		t.Errorf("RandomUUID() returned ZeroUUID")
	}

	// Test that a random UUID is not equal to another random UUID.
	uuid2 := randomUUID()
	if uuid == uuid2 {
		t.Errorf("RandomUUID() returned the same UUID twice")
	}

	// Test that the base64 URL-safe representation of a random UUID does not begin with a hyphen.
	uuidString := uuid.String()
	if strings.HasPrefix(uuidString, "-") {
		t.Errorf("RandomUUID() returned a UUID with a hyphen-prefixed base64 URL-safe representation")
	}
}
