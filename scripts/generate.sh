#!/usr/bin/env bash

#
# Copyright (c) 2023 Asim Ihsan.
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.
#
# SPDX-License-Identifier: MPL-2.0
#

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT_DIR="$( cd "$SCRIPT_DIR/.." >/dev/null 2>&1 && pwd )"

pushd $ROOT_DIR || exit > /dev/null
trap "popd > /dev/null" EXIT

# Build the Docker image
docker build -t my-antlr-build .

# Generate a unique container name with a timestamp
container_name="temp-container-$(date +%s)"

# Create a temporary container
docker create --name "$container_name" my-antlr-build

# Copy the generated ANTLR Go target files to the local machine
rm -rf ./generated
docker cp "$container_name":/app/antlr/generated ./

# Remove the temporary container
docker rm "$container_name"

echo "ANTLR Go target files have been copied to ./generated"
