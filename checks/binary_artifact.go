// Copyright 2021 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package checks

import (
	"strings"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/checks/evaluation"
	"github.com/ossf/scorecard/v5/checks/raw"
	"github.com/ossf/scorecard/v5/clients"
	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/internal/packageclient"
	sclog "github.com/ossf/scorecard/v5/log"
	"github.com/ossf/scorecard/v5/probes"
	"github.com/ossf/scorecard/v5/probes/zrunner"
)

// CheckBinaryArtifacts is the exported name for Binary-Artifacts check.
const CheckBinaryArtifacts string = "Binary-Artifacts"
const selfLabel string = "SELF"

//nolint:gochecknoinits
func init() {
	supportedRequestTypes := []checker.RequestType{
		checker.CommitBased,
		checker.FileBased,
	}
	if err := registerCheck(CheckBinaryArtifacts, BinaryArtifacts, supportedRequestTypes); err != nil {
		// this should never happen
		panic(err)
	}
}

// BinaryArtifacts  will check the repository contains binary artifacts.
func BinaryArtifacts(c *checker.CheckRequest) checker.CheckResult {

	rawData, err := raw.BinaryArtifacts(c)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckBinaryArtifacts, e)
	}

	// Set the raw results.
	pRawResults := getRawResults(c)
	pRawResults.BinaryArtifactResults = rawData

	// Evaluate the probes.
	findings, err := zrunner.Run(pRawResults, probes.BinaryArtifacts)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckBinaryArtifacts, e)
	}

	ret := evaluation.BinaryArtifacts(CheckBinaryArtifacts, findings, c.Dlogger)
	ret.Findings = findings

	BinaryArtifactsDependencies(c)

	return ret
}

// BinaryArtifactsDependencies will check all depdencies of repository contains binary artifacts and log all that are found.
func BinaryArtifactsDependencies(c *checker.CheckRequest) bool {

	// if package name wasn't given on the command line, try to find it using the repo url
	if c.ProjectClient.GetPackageName() == "" || c.ProjectClient.GetSystem() == "" {

		// Gets system
		uriComponents := strings.Split(c.RepoClient.URI(), "/")
		host := uriComponents[0]
		project := uriComponents[1] + "/" + uriComponents[2]
		versions, err := c.ProjectClient.GetProjectPackageVersions(c.Ctx, host, project)
		if err != nil {
			return false
		}
		system := versions.Versions[0].VersionKey.System

		// Repos are often mapped to by multiple package names
		// Therefore, only include packages that have the same name as the repo url (ex. most GO packages)
		// Doing this instead of VersionKey.Name gets rid of most false
		// positive matches but will cause some false negatives

		c.ProjectClient = packageclient.CreateDepsDevClientForPackage(c.RepoClient.URI(), system)
	}

	dependencies, err := c.ProjectClient.GetPackageDependencies(c.Ctx)
	if err != nil {
		return false
	}
	logger := sclog.NewLogger(sclog.DefaultLevel)
	numSkipped := 0 // do something with this eventually?

	for _, dep := range dependencies.Nodes {
		if dep.Relation == selfLabel {
			continue
		}
		depURI, err := c.ProjectClient.GetURI(c.Ctx, dep.VersionKey.Name, dep.VersionKey.Version, dep.VersionKey.System)
		if err != nil {
			numSkipped++
			continue
		}

		repoClient := c.ProjectClient.CreateGithubRepoClient(c.Ctx, logger)
		repo, _, _, _, _, _, err := checker.GetClients(c.Ctx, depURI, "", "", "", logger) // change this?
		if err != nil {
			numSkipped++
			continue
		}
		err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
		if err != nil {
			numSkipped++
			continue
		}
		dc := checker.CheckRequest{
			Ctx:        c.Ctx,
			RepoClient: repoClient,
			Repo:       repo,
			Dlogger:    c.Dlogger,
		}

		depRawData, err := raw.BinaryArtifacts(&dc)
		if err != nil {
			continue
		}

		// Set the raw results.
		dRawResults := getRawResults(c)
		dRawResults.BinaryArtifactResults = depRawData

		// Evaluate the probes.
		findings, err := zrunner.Run(dRawResults, probes.BinaryArtifacts)
		if err != nil {
			continue
		}

		// log
		evaluation.BinaryArtifactsDependencies(CheckBinaryArtifacts, dep.VersionKey.Name, findings, dc.Dlogger)
	}

	return true

}
