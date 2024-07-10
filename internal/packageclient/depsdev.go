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

package packageclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/githubrepo"
	"github.com/ossf/scorecard/v5/clients/gitlabrepo"
	"github.com/ossf/scorecard/v5/log"
)

var (
	sourceRepoLabel = "SOURCE_REPO"
	githubDomain    = regexp.MustCompile("github.com/.*")
)

// This interface lets Scorecard look up package manager metadata for a project.
type ProjectPackageClient interface {
	GetProjectPackageVersions(ctx context.Context, host, project string) (*ProjectPackageVersions, error)
	GetPackage(ctx context.Context) (*PackageData, error)
	GetPackageDependencies(ctx context.Context) (*PackageDependencies, error)
	GetVersion(ctx context.Context, name, version, system string) (*VersionData, error)
	GetURI(ctx context.Context, name, version, system string) (string, error)
	GetPackageName() string
	GetSystem() string
	CreateGitlabRepoClient(ctx context.Context, host string) (clients.RepoClient, error)
	CreateGithubRepoClient(ctx context.Context, l *log.Logger) clients.RepoClient
}

type depsDevClient struct {
	client      *http.Client
	packageName string
	system      string
}

type ProjectPackageVersions struct {
	// field alignment
	//nolint:govet
	Versions []struct {
		VersionKey struct {
			System  string `json:"system"`
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"versionKey"`
		SLSAProvenances []struct {
			SourceRepository string `json:"sourceRepository"`
			Commit           string `json:"commit"`
			Verified         bool   `json:"verified"`
		} `json:"slsaProvenances"`
		RelationType       string `json:"relationType"`
		RelationProvenance string `json:"relationProvenance"`
	} `json:"versions"`
}

type PackageData struct {
	PackageKey struct {
		System string `json:"system"`
		Name   string `json:"name"`
	} `json:"packageKey"`
	Versions []struct {
		VersionKey struct {
			System  string `json:"system"`
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"versionKey"`
		Purl        string `json:"purl"`
		PublishedAt string `json:"publishedAt"`
		IsDefault   bool   `json:"isDefault"`
	} `json:"versions"`
}

type PackageDependencies struct {
	Nodes []struct {
		VersionKey struct {
			System  string `json:"system"`
			Name    string `json:"name"`
			Version string `json:"version"`
		}
		Bundled  bool     `json:"bundled"`
		Relation string   `json:"relation"`
		Errors   []string `json:"errors"`
	} `json:"nodes"`
	Edges []struct {
		FromNode    int    `json:"fromNode"`
		ToNode      int    `json:"toNode"`
		Requirement string `json:"requirement"`
	} `json:"edges"`
	Error string `json:"error"`
}

type VersionData struct {
	VersionKey struct {
		System  string `json:"system"`
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"versionKey"`
	Purl         string   `json:"purl"`
	PublishedAt  string   `json:"publishedAt"`
	IsDefault    bool     `json:"isDefault"`
	Licenses     []string `json:"licenses"`
	AdvisoryKeys []any    `json:"advisoryKeys"`
	Links        []struct {
		Label string `json:"label"`
		URL   string `json:"url"`
	} `json:"links"`
}

func CreateDepsDevClient() ProjectPackageClient {
	return depsDevClient{
		client: &http.Client{},
	}
}

func CreateDepsDevClientForPackage(packageName, system string) ProjectPackageClient {
	return depsDevClient{
		client:      &http.Client{},
		packageName: packageName,
		system:      system,
	}
}

var (
	ErrDepsDevAPI            = errors.New("deps.dev")
	ErrProjNotFoundInDepsDev = errors.New("project not found in deps.dev")
	ErrPkgNotFoundInDepsDev  = errors.New("package not found in deps.dev")
)

func (d depsDevClient) GetProjectPackageVersions(
	ctx context.Context, host, project string,
) (*ProjectPackageVersions, error) {
	path := fmt.Sprintf("%s/%s", host, project)
	query := fmt.Sprintf("https://api.deps.dev/v3/projects/%s:packageversions", url.QueryEscape(path))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, query, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("deps.dev GetProjectPackageVersions: %w", err)
	}
	defer resp.Body.Close()

	var res ProjectPackageVersions
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrProjNotFoundInDepsDev
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrDepsDevAPI, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("resp.Body.Read: %w", err)
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, fmt.Errorf("deps.dev json.Unmarshal: %w", err)
	}

	return &res, nil
}

func (d depsDevClient) GetPackageDependencies(
	ctx context.Context) (*PackageDependencies, error) {

	// GetPackage used to get the default version. Requires the system to be specified so
	// this call must be done after GetProjectPackageVersions
	packageInfo, err := d.GetPackage(ctx)
	if err != nil {
		return nil, fmt.Errorf("deps.dev GetPackage: %w", err)
	}

	defaultVersion := packageInfo.Versions[0].VersionKey.Version
	for _, ver := range packageInfo.Versions {
		if ver.IsDefault {
			defaultVersion = ver.VersionKey.Version
			break
		}
	}

	query := fmt.Sprintf("https://api.deps.dev/v3alpha/systems/%s/packages/%s/versions/%s:dependencies",
		url.QueryEscape(d.GetSystem()), url.QueryEscape(d.GetPackageName()), url.QueryEscape(defaultVersion))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, query, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("deps.dev GetPackageDependencies: %w", err)
	}
	defer resp.Body.Close()

	var res PackageDependencies

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrPkgNotFoundInDepsDev
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrDepsDevAPI, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("resp.Body.Read: %w", err)
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, fmt.Errorf("deps.dev json.Unmarshal: %w", err)
	}
	return &res, nil
}

func (d depsDevClient) GetPackage(
	ctx context.Context) (*PackageData, error) {

	query := fmt.Sprintf("https://api.deps.dev/v3alpha/systems/%s/packages/%s",
		url.QueryEscape(d.GetSystem()), url.QueryEscape(d.GetPackageName()))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, query, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("deps.dev GetPackage: %w", err)
	}
	defer resp.Body.Close()

	var res PackageData
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrPkgNotFoundInDepsDev
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrDepsDevAPI, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("resp.Body.Read: %w", err)
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, fmt.Errorf("deps.dev json.Unmarshal: %w", err)
	}
	return &res, nil
}

func (d depsDevClient) GetVersion(
	ctx context.Context, name, version, system string) (*VersionData, error) {

	query := fmt.Sprintf("https://api.deps.dev/v3alpha/systems/%s/packages/%s/versions/%s", url.QueryEscape(system),
		url.QueryEscape(name), url.QueryEscape(version))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, query, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("deps.dev GetVersion: %w", err)
	}
	defer resp.Body.Close()

	var res VersionData
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrPkgNotFoundInDepsDev
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrDepsDevAPI, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("resp.Body.Read: %w", err)
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, fmt.Errorf("deps.dev json.Unmarshal: %w", err)
	}
	return &res, nil
}

func (d depsDevClient) GetURI(
	ctx context.Context, name, version, system string,
) (string, error) {
	versionInfo, err := d.GetVersion(ctx, name, version, system)
	if err != nil {
		return "", fmt.Errorf("deps.dev GetVersion: %s", name)
	}
	trimmedURL := ""
	for _, ver := range versionInfo.Links {
		if ver.Label == sourceRepoLabel {
			trimmedURL = strings.TrimSuffix(ver.URL, ".git")
			trimmedURL = githubDomain.FindString(trimmedURL)
			break
		}
	}
	if trimmedURL == "" {
		return "", fmt.Errorf("deps.dev GetURI: %s", name)
	}
	return trimmedURL, nil
}

func (d depsDevClient) GetPackageName() string {
	return d.packageName
}

func (d depsDevClient) GetSystem() string {
	return d.system
}

func (d depsDevClient) CreateGithubRepoClient(ctx context.Context, l *log.Logger) clients.RepoClient {
	return githubrepo.CreateGithubRepoClient(ctx, l)
}

func (d depsDevClient) CreateGitlabRepoClient(ctx context.Context, host string) (clients.RepoClient, error) {
	ret, err := gitlabrepo.CreateGitlabClient(ctx, host)
	return ret, err
}
