// Copyright 2024 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ossf/scorecard/v5/internal/packageclient"
)

var _ = Describe("E2E TEST: depsdevclient.GetProjectPackageVersions", func() {
	var client packageclient.ProjectPackageClient

	Context("E2E TEST: Confirm ProjectPackageClient works", func() {
		It("Should receive a non-empty response from deps.dev for existing projects", func() {
			client = packageclient.CreateDepsDevClient()
			versions, err := client.GetProjectPackageVersions(
				context.Background(), "github.com", "ossf/scorecard",
			)
			Expect(err).Should(BeNil())
			Expect(len(versions.Versions)).Should(BeNumerically(">", 0))
		})
		It("Should error from deps.dev for nonexistent projects", func() {
			client = packageclient.CreateDepsDevClient()
			versions, err := client.GetProjectPackageVersions(
				context.Background(), "github.com", "ossf/scorecard-E2E-TEST-DOES-NOT-EXIST",
			)
			Expect(err).ShouldNot(BeNil())
			Expect(versions).Should(BeNil())
		})
		It("Should receive a non-empty response from deps.dev for existing projects", func() {
			client = packageclient.CreateDepsDevClient()
			versions, err := client.GetProjectPackageVersions(
				context.Background(), "gitlab.com", "libtiff/libtiff",
			)
			Expect(err).Should(BeNil())
			Expect(len(versions.Versions)).Should(BeNumerically(">", 0))
		})
	})
})

var _ = Describe("E2E TEST: depsdevclient.GetPackage", func() {
	var client packageclient.ProjectPackageClient

	Context("E2E TEST: Confirm GetPackage works", func() {
		It("Should receive a non-empty response from deps.dev for existing projects", func() {
			client = packageclient.CreateDepsDevClient()
			versions, err := client.GetPackage(
				context.Background(), "github.com", "ossf/scorecard", "GO",
			)
			Expect(err).Should(BeNil())
			Expect(len(versions.Versions)).Should(BeNumerically(">", 0))
		})
		It("Should error from deps.dev for nonexistent projects", func() {
			client = packageclient.CreateDepsDevClient()
			versions, err := client.GetPackage(
				context.Background(), "github.com", "ossf/scorecard-E2E-TEST-DOES-NOT-EXIST", "GO",
			)
			Expect(err).ShouldNot(BeNil())
			Expect(versions).Should(BeNil())
		})
	})
})

var _ = Describe("E2E TEST: depsdevclient.GetPackageDependencies", func() {
	var client packageclient.ProjectPackageClient

	Context("E2E TEST: Confirm GetPackageDependencies works", func() {
		It("Should receive a non-empty response from deps.dev for existing projects", func() {
			client = packageclient.CreateDepsDevClient()
			dependencies, err := client.GetPackageDependencies(
				context.Background(), "github.com", "ossf/scorecard",
			)
			Expect(err).Should(BeNil())
			Expect(len(dependencies.Nodes)).Should(BeNumerically(">", 0))
		})
		It("Should error from deps.dev for nonexistent projects", func() {
			client = packageclient.CreateDepsDevClient()
			versions, err := client.GetPackageDependencies(
				context.Background(), "github.com", "ossf/scorecard-E2E-TEST-DOES-NOT-EXIST",
			)
			Expect(err).ShouldNot(BeNil())
			Expect(versions).Should(BeNil())
		})
	})
})

var _ = Describe("E2E TEST: depsdevclient.GetVersion", func() {
	var client packageclient.ProjectPackageClient

	Context("E2E TEST: Confirm ProjectPackageClient works", func() {
		It("Should receive a non-empty response from deps.dev for existing projects", func() {
			client = packageclient.CreateDepsDevClient()
			versionData, err := client.GetVersion(
				context.Background(), "github.com/ossf/scorecard", "v1.2.0", "GO",
			)
			Expect(err).Should(BeNil())
			Expect(versionData.VersionKey.Version).Should(Equal("v1.2.0"))
		})
		It("Should error from deps.dev for nonexistent projects", func() {
			client = packageclient.CreateDepsDevClient()
			versions, err := client.GetVersion(
				context.Background(), "github.com/ossf/scorecard-E2E-TEST-DOES-NOT-EXIST", "v2.4.3", "GO",
			)
			Expect(err).ShouldNot(BeNil())
			Expect(versions).Should(BeNil())
		})
	})
})

var _ = Describe("E2E TEST: depsdevclient.GetURI", func() {
	var client packageclient.ProjectPackageClient

	Context("E2E TEST: Confirm GetURI works", func() {
		It("Should error from deps.dev for nonexistent projects", func() {
			client = packageclient.CreateDepsDevClient()
			URI, err := client.GetURI(
				context.Background(), "github.com/ossf/scorecard-E2E-TEST-DOES-NOT-EXIST", "v2.4.3", "GO",
			)
			Expect(err).ShouldNot(BeNil())
			Expect(URI).Should(Equal(""))
		})
		It("Should receive a non-empty response from deps.dev for existing projects", func() {
			client = packageclient.CreateDepsDevClient()
			URI, err := client.GetURI(
				context.Background(), "@colors/colors", "1.5.0", "NPM",
			)
			Expect(err).Should(BeNil())
			Expect(URI).Should(Equal("github.com/DABH/colors.js"))
		})
		It("Should error from deps.dev for non-github url", func() {
			client = packageclient.CreateDepsDevClient()
			URI, err := client.GetURI(
				context.Background(), "golang.org/x/crypto", "v0.24.0", "GO",
			)
			Expect(err).ShouldNot(BeNil())
			Expect(URI).Should(Equal(""))
		})
	})
})
