/*
 * Copyright 2019 The original author or authors
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
 */

package testing

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

type Testcases struct {
	Common    Testcase   `toml:"common"`
	Testcases []Testcase `toml:"testcases"`
}

func (tcs *Testcases) Run(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	for _, tc := range tcs.Testcases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.merge(&tcs.Common).Run(t)
		})
	}
}

type Testcase struct {
	Name        string `toml:"name"`
	Repo        string `toml:"repo"`
	Refspec     string `toml:"refspec"`
	SubPath     string `toml:"sub-path"`
	Artifact    string `toml:"artifact"`
	Handler     string `toml:"handler"`
	Override    string `toml:"override"`
	ContentType string `toml:"content-type"`
	Input       string `toml:"input"`
	Output      string `toml:"output"`
	SkipRebuild bool   `toml:"skip-rebuild"`
}

func (tc *Testcase) Run(t *testing.T) {
	appdir, err := ioutil.TempDir("", "builder-")
	if err != nil {
		t.Fatalf("could not create temp dir: %v", err)
	} else {
		defer func() { _ = os.RemoveAll(appdir) }()
	}
	fndir := filepath.Join(appdir, tc.SubPath)

	tc.cloneRepo(t, appdir)

	lastSlash := strings.LastIndex(tc.Repo, "/")
	fnImg := fmt.Sprintf("builder-tests/%s-%d", tc.Repo[lastSlash+1:], rand.Int31n(10000))

	t.Run("build", func(t *testing.T) {
		tc.createFunctionImg(t, fnImg, fndir)
	})

	t.Run("run", func(t *testing.T) {
		localPort, docker := tc.startServer(t, fnImg)
		tc.invokeFunction(t, localPort)
		tc.stopFunctionContainer(t, docker)
	})

	if !tc.SkipRebuild {
		t.Run("rebuild", func(t *testing.T) {
			// Re-create function, should use cache
			t.Run("build", func(t *testing.T) {
				tc.createFunctionImg(t, fnImg, fndir)
			})
			t.Run("run", func(t *testing.T) {
				localPort2, docker := tc.startServer(t, fnImg)
				tc.invokeFunction(t, localPort2)
				tc.stopFunctionContainer(t, docker)
			})
		})
	}

	tc.deleteImage(t, fnImg)
}

func (tc *Testcase) merge(c *Testcase) *Testcase {
	if tc.Repo == "" {
		tc.Repo = c.Repo
	}
	if tc.Refspec == "" {
		tc.Refspec = c.Refspec
	}
	if tc.SubPath == "" {
		tc.SubPath = c.SubPath
	}
	if tc.Artifact == "" {
		tc.Artifact = c.Artifact
	}
	if tc.Handler == "" {
		tc.Handler = c.Handler
	}
	if tc.Override == "" {
		tc.Override = c.Override
	}
	if tc.ContentType == "" {
		tc.ContentType = c.ContentType
	}
	if tc.Input == "" {
		tc.Input = c.Input
	}
	if tc.Output == "" {
		tc.Output = c.Output
	}

	return tc
}

func (tc *Testcase) deleteImage(t *testing.T, fnImg string) {
	if err := tc.runCmd("docker", "rmi", "--force", fnImg); err != nil {
		t.Fatalf("could not remove image %q: %v", fnImg, err)
	}
}

func (tc *Testcase) invokeFunction(t *testing.T, localPort int32) {
	if resp, err := http.Post(fmt.Sprintf("http://localhost:%d", localPort), tc.ContentType, strings.NewReader(tc.Input)); err != nil {
		t.Fatalf("could not post to function: %v", err)
	} else {
		if result, err := ioutil.ReadAll(resp.Body); err != nil {
			t.Fatalf("could not read response from function: %v", err)
		} else if string(result) != tc.Output {
			t.Fatalf("unexpected result from function: %q != %q", string(result), tc.Output)
		}
	}
}

func (tc *Testcase) stopFunctionContainer(t *testing.T, docker *exec.Cmd) {
	if err := docker.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("could not kill app: %v", err)
	}
}

func (tc *Testcase) startServer(t *testing.T, fnImg string) (int32, *exec.Cmd) {
	localPort := 1024 + rand.Int31n(65535-1024)
	var docker *exec.Cmd
	docker, err := tc.startCmd("docker", "run", "-p", fmt.Sprintf("%d:8080", localPort), fnImg)
	if err != nil {
		t.Fatalf("could not run built function: %v", err)
	}
	addr := fmt.Sprintf("http://localhost:%d", localPort)

	until := time.Now().Add(20 * time.Second)
	for ; time.Now().Before(until); time.Sleep(1 * time.Second) {
		_, err := http.Get(addr)
		if err == nil {
			break
		}
		fmt.Printf("Could not connect to %s, retrying...\n", addr)
	}

	return localPort, docker
}

func (tc *Testcase) createFunctionImg(t *testing.T, fnImg string, appdir string) {
	err := tc.runCmd("pack", "build", "--no-pull",
		"--builder", "projectriff/builder",
		"--path", appdir,
		"--env", fmt.Sprintf("%s=%s", "RIFF", "true"),
		"--env", fmt.Sprintf("%s=%s", "RIFF_ARTIFACT", tc.Artifact),
		"--env", fmt.Sprintf("%s=%s", "RIFF_HANDLER", tc.Handler),
		"--env", fmt.Sprintf("%s=%s", "RIFF_OVERRIDE", tc.Override),
		fnImg)
	if err != nil {
		t.Fatalf("could not build: %v", err)
	}
}

func (tc *Testcase) cloneRepo(t *testing.T, appdir string) {
	if err := tc.runCmd("git", "clone", tc.Repo, appdir); err != nil {
		t.Fatalf("could not clone into %q: %v", appdir, err)
	}
	if tc.Refspec != "" {
		dir, _ := os.Getwd()
		defer os.Chdir(dir)
		os.Chdir(appdir)
		if err := tc.runCmd("git", "checkout", tc.Refspec); err != nil {
			t.Fatalf("could not checkout %q: %v", tc.Refspec, err)
		}
	}
}

func (tc *Testcase) runCmd(c string, s ...string) error {
	if cmd, err := tc.startCmd(c, s...); err != nil {
		return err
	} else {
		return cmd.Wait()
	}
}

func (tc *Testcase) startCmd(c string, s ...string) (*exec.Cmd, error) {
	fmt.Printf("RUNNING %s %s\n", c, strings.Join(s, " "))
	command := exec.Command(c, s...)
	command.Stderr = os.Stderr
	command.Stdout = os.Stdout
	return command, command.Start()
}
