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

// DependencyClient is used for creating RepoClients for dependencies of a package.
// Acts as a wrapper for githubrepo.CreateGithubRepoClient() to allow for mocking the response in unit tests.
package dependencyclient

import (
	"context"
	"net/http"

	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/githubrepo"
	"github.com/ossf/scorecard/v5/clients/gitlabrepo"
	"github.com/ossf/scorecard/v5/log"
)

type DependencyClient interface {
	CreateGithubRepoClient(context.Context, *log.Logger) clients.RepoClient
	CreateGitlabRepoClient(context.Context, string) clients.RepoClient
}

type depClient struct {
	client *http.Client
}

func CreateDependencyClient() DependencyClient {
	return depClient{
		client: &http.Client{},
	}
}

func (d depClient) CreateGithubRepoClient(ctx context.Context, l *log.Logger) clients.RepoClient {
	return githubrepo.CreateGithubRepoClient(ctx, l)
}

func (d depClient) CreateGitlabRepoClient(ctx context.Context, host string) clients.RepoClient {
	ret, _ := gitlabrepo.CreateGitlabClient(ctx, host)
	return ret
}
