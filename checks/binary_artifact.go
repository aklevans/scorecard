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
	sclog "github.com/ossf/scorecard/v5/log"
	"github.com/ossf/scorecard/v5/probes"
	"github.com/ossf/scorecard/v5/probes/zrunner"
)

// CheckBinaryArtifacts is the exported name for Binary-Artifacts check.
const CheckBinaryArtifacts string = "Binary-Artifacts"

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
	return ret
}

// BinaryArtifactsDependencies will check all depdencies of repository contains binary artifacts.
func BinaryArtifactsDependencies(c *checker.CheckRequest) checker.CheckResult {
	uriComponents := strings.Split(c.Repo.URI(), "/")
	host := uriComponents[0]
	project := uriComponents[1] + "/" + uriComponents[2]
	dependencies, err := c.ProjectClient.GetPackageDependencies(c.Ctx, host, project)
	if err != nil {
		e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		return checker.CreateRuntimeErrorResult(CheckBinaryArtifacts, e)
	}
	rawData := checker.BinaryArtifactData{}
	logger := sclog.NewLogger(sclog.DefaultLevel)
	numSkipped := 0

	// todo: self is currently included in dependency list. Exclude?
	for _, dep := range dependencies.Nodes {
		depURI, err := c.ProjectClient.GetURI(c.Ctx, dep.VersionKey.Name, dep.VersionKey.Version, dep.VersionKey.System)
		if err != nil {
			numSkipped += 1
			continue // if cant find github url for dependency, skip for now
		}

		repo, repoClient, _, _, _, _, err := checker.GetClients(c.Ctx, depURI, "", logger)
		if err != nil {
			e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
			return checker.CreateRuntimeErrorResult(CheckBinaryArtifacts, e)
		}
		err = repoClient.InitRepo(repo, clients.HeadSHA, 0)
		if err != nil {
			e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
			return checker.CreateRuntimeErrorResult(CheckBinaryArtifacts, e)
		}
		dc := checker.CheckRequest{
			Ctx:        c.Ctx,
			RepoClient: repoClient,
			Repo:       repo,
			Dlogger:    c.Dlogger,
		}

		depRawData, err := raw.BinaryArtifacts(&dc)
		if err != nil {
			e := sce.WithMessage(sce.ErrScorecardInternal, err.Error())
			return checker.CreateRuntimeErrorResult(CheckBinaryArtifacts, e)
		}

		rawData.Files = append(rawData.Files, depRawData.Files...)
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
	return ret
}
