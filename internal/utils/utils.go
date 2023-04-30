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

// The value is either a quoted string with escaped characters, or an unescaped string without quotes.
// We need to strip the quotes and unescape the string.
func HandleStringLiteral(item string) string {
	if len(item) == 0 {
		return ""
	}
	if item[0] == '"' {
		unquoted := item[1 : len(item)-1]
		for i := 0; i < len(unquoted); i++ {
			if unquoted[i] == '\\' {
				if i == len(unquoted)-1 {
					return unquoted
				}
				switch unquoted[i+1] {
				case 'n':
					unquoted = unquoted[:i] + "\n" + unquoted[i+2:]
				case 'r':
					unquoted = unquoted[:i] + "\r" + unquoted[i+2:]
				case 't':
					unquoted = unquoted[:i] + "\t" + unquoted[i+2:]
				case '\\':
					unquoted = unquoted[:i] + "\\" + unquoted[i+2:]
				case '"':
					unquoted = unquoted[:i] + "\"" + unquoted[i+2:]
				default:
					return unquoted
				}
			}
		}
		return unquoted
	}
	return item
}
