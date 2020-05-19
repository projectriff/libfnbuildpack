/*
 * Copyright 2018-2020 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package libfnbuildpack

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/libpak"
)

// IsRiff determines if an application is explicitly a riff application.  This can be indicated by setting the $RIFF
// environment variable, or having a `<path>/riff.toml` file.
func IsRiff(path string, configurationResolver libpak.ConfigurationResolver) (bool, error) {
	file := filepath.Join(path, "riff.toml")
	if _, err := os.Stat(file); err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("unable to stat %s\n%w", file, err)
	} else if err == nil {
		return true, nil
	}

	if _, ok := configurationResolver.Resolve("RIFF"); ok {
		return true, nil
	}

	return false, nil
}

// Metadata returns metadata about a riff application beginning with the contents of `riff.toml` and then overriding
// with any contents set in `$RIFF_ARTIFACT` and `$RIFF_HANDLER`.
func Metadata(path string, configurationResolver libpak.ConfigurationResolver) (map[string]interface{}, error) {
	metadata := make(map[string]interface{})

	file := filepath.Join(path, "riff.toml")
	if _, err := toml.DecodeFile(file, &metadata); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("unable to decode %s\n%w", file, err)
	}

	if s, ok := configurationResolver.Resolve("RIFF_ARTIFACT"); ok {
		metadata["artifact"] = s
	}

	if s, ok := configurationResolver.Resolve("RIFF_HANDLER"); ok {
		metadata["handler"] = s
	}

	return metadata, nil
}
