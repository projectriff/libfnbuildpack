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

package node

import (
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/nodejs-cnb/node"
	"github.com/projectriff/riff-buildpack/pkg/metadata"
)

func BuildPlanContribution(detect detect.Detect, metadata metadata.Metadata) buildplan.BuildPlan {
	n := detect.BuildPlan[node.Dependency]
	if n.Metadata == nil {
		n.Metadata = buildplan.Metadata{}
	}
	n.Metadata["launch"] = true
	n.Metadata["build"] = true

	r := detect.BuildPlan[Dependency]
	if r.Metadata == nil {
		r.Metadata = buildplan.Metadata{}
	}
	r.Metadata[FunctionArtifact] = metadata.Artifact

	return buildplan.BuildPlan{node.Dependency: n, Dependency: r}
}
