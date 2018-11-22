#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

gcloud auth activate-service-account --key-file <(echo $GCLOUD_CLIENT_SECRET | base64 --decode)

gsutil cp -a public-read scratch/io/projectriff/riff/io.projectriff.riff/*/*.tgz gs://projectriff/riff-buildpack/
gsutil cp -a public-read scratch/io/projectriff/riff/io.projectriff.riff/*/*.tgz gs://projectriff/riff-buildpack/latest.tgz