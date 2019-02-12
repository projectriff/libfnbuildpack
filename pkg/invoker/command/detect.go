/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package command

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/projectriff/riff-buildpack/pkg/metadata"
)

func DetectCommand(detect detect.Detect, metadata metadata.Metadata) (bool, error) {
	if metadata.Artifact == "" {
		return false, nil
	}

	path := filepath.Join(detect.Application.Root, metadata.Artifact)

	ok, err := helper.FileExists(path)
	if err != nil || !ok {
		return false, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if info.Mode().IsRegular() && info.Mode().Perm()&0100 == 0100 {
		return true, nil
	} else {
		detect.Logger.Debug("Disregarding %q for the 'command' invoker, as it does not have executable permission", path)
		return false, nil
	}
}
