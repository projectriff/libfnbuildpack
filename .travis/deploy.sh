#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

./deploy.sh
./trigger-builder-build.sh
