/*
 * Copyright 2018 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package java

import (
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/openjdk-buildpack/jre"
	"github.com/projectriff/riff-buildpack/pkg/metadata"
)

// BuildPlanContribution returns the BuildPlan with requirements for the invoker
func BuildPlanContribution(detect detect.Detect, metadata metadata.Metadata) buildplan.BuildPlan {
	j := detect.BuildPlan[jre.Dependency]
	if j.Metadata == nil {
		j.Metadata = buildplan.Metadata{}
	}
	j.Metadata[jre.LaunchContribution] = true

	r := detect.BuildPlan[Dependency]
	if r.Metadata == nil {
		r.Metadata = buildplan.Metadata{}
	}
	r.Metadata[Handler] = metadata.Handler

	return buildplan.BuildPlan{jre.Dependency: j, Dependency: r}
}
