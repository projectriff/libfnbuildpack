/*
 * Copyright 2018-2019 the original author or authors.
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

package function

import (
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/detect"
)

// BuildpackImplementation is an interface for types that implement concrete behaviors for build and detect.
type BuildpackImplementation interface {
	// Build is called during the build phase of the buildpack lifecycle.
	Build(build build.Build) (int, error)

	// Detect is called during the detect phase of the buildpack lifecycle.
	Detect(detect detect.Detect, metadata Metadata) (int, error)

	// Id returns the id of the function buildpack.
	Id() string
}
