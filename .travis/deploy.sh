#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

source `dirname "${BASH_SOURCE[0]}"`/deploy.sh
source `dirname "${BASH_SOURCE[0]}"`/trigger-builder-build.sh
