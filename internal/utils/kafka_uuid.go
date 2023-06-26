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

/// See: https://github.com/apache/kafka/blob/c5889fceddb9a0174452ae60a57c8ff3f087a6a4/clients/src/main/java/org/apache/kafka/common/Uuid.java#L28

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"strings"
)

// UUID represents a universally unique identifier (UUID).
type UUID struct {
	mostSignificantBits  uint64
	leastSignificantBits uint64
}

// MetadataTopicID is a UUID for the metadata topic in KRaft mode.
// It will never be returned by the randomUUID function.
var MetadataTopicID = UUID{0, 1}

// ZeroUUID represents a null or empty UUID.
// It will never be returned by the randomUUID function.
var ZeroUUID = UUID{0, 0}

// randomUUID generates a type 4 (pseudo randomly generated) UUID
// with the following constraints:
// - can't be most == 0 and least == 1
// - can't be most == 0 and least == 0
// - the base64 URL-safe representation can't begin with a hyphen
func randomUUID() UUID {
	uuid := unsafeRandomUUID()
	for uuid == MetadataTopicID || uuid == ZeroUUID || strings.HasPrefix(uuid.String(), "-") {
		uuid = unsafeRandomUUID()
	}
	return uuid
}

// unsafeRandomUUID generates a type 4 (pseudo randomly generated) UUID
// without any constraints.
func unsafeRandomUUID() UUID {
	mostSignificantBits := randomUint64()
	leastSignificantBits := randomUint64()
	return UUID{mostSignificantBits, leastSignificantBits}
}

// randomUint64 generates a random uint64 value.
func randomUint64() uint64 {
	buf := make([]byte, 8)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
	return binary.BigEndian.Uint64(buf)
}

// String returns a base64 URL-safe encoding of the UUID without padding.
func (u UUID) String() string {
	bytes := u.getBytesFromUUID()
	return base64.RawURLEncoding.EncodeToString(bytes)
}

// getBytesFromUUID extracts bytes for the UUID, which is 128 bits (or 16 bytes) long.
func (u UUID) getBytesFromUUID() []byte {
	buf := make([]byte, 16)
	binary.BigEndian.PutUint64(buf[:8], u.mostSignificantBits)
	binary.BigEndian.PutUint64(buf[8:], u.leastSignificantBits)
	return buf
}
