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

# Build the binary
docker buildx build -f Dockerfile.binary -t virtual-cluster-build --load .

# Generate a unique container name with a timestamp
container_name="temp-container-$(date +%s)"

# Create a temporary container
docker create --name "$container_name" virtual-cluster-build

# Copy the binary to the local machine
rm -f ./build/virtual-cluster
docker cp "$container_name":/app/build/virtual-cluster ./build/virtual-cluster

# Remove the temporary container
docker rm "$container_name"

echo "Binary has been copied to ./build/virtual-cluster"
