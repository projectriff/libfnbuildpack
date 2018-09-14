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
	"github.com/buildpack/libbuildpack"
	"github.com/projectriff/riff-buildpack"
)

const (
	// Handler is the key identifying the riff handler metadata in the build plan
	Handler = "handler"

	// RiffInvokerDependency is the key identifying the riff invoker in the buildpack plan.
	RiffInvokerDependency = "riff-invoker-java"
)

// RiffInvoker represents the Java invoker contributed by the buildpack
type RiffInvoker struct {
}

// BuildPlanContribution returns the BuildPlan with requirements for the invoker
func (r RiffInvoker) BuildPlanContribution(metadata riff_buildpack.Metadata) libbuildpack.BuildPlan {
	// TODO use constants for openjdk-jre and launch
	return libbuildpack.BuildPlan{
		"openjdk-jre": libbuildpack.BuildPlanDependency{
			Metadata: libbuildpack.BuildPlanDependencyMetadata{"launch": true},
			Version:  "1.*",
		},
		RiffInvokerDependency: libbuildpack.BuildPlanDependency{
			Metadata: libbuildpack.BuildPlanDependencyMetadata{
				Handler: metadata.Handler,
			},
		},
	}
}
