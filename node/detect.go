/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package node

import (
	"github.com/cloudfoundry/libjavabuildpack"
	"github.com/projectriff/riff-buildpack"
	"path/filepath"
)

// DetectNode answers true if the `artifact` path is set, the file exists and ends in ".js"
func DetectNode(detect libjavabuildpack.Detect, metadata riff_buildpack.Metadata) (bool, error) {
	if metadata.Artifact == "" {
		return false, nil
	}

	path := filepath.Join(detect.Application.Root, metadata.Artifact)

	ok, err := libjavabuildpack.FileExists(path)
	if err != nil || !ok {
		return false, err
	}
	return filepath.Ext(path) == ".js", nil
}
