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
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/githubrepo"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	"github.com/ossf/scorecard/v5/internal/packageclient"
	sclog "github.com/ossf/scorecard/v5/log"
	scut "github.com/ossf/scorecard/v5/utests"
)

func TestBinaryArtifacts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		inputFolder string
		err         error
		expected    scut.TestReturn
	}{
		{
			name:        "Jar file",
			inputFolder: "testdata/binaryartifacts/jars",
			err:         nil,
			expected: scut.TestReturn{
				Score:        8,
				NumberOfInfo: 0,
				NumberOfWarn: 2,
			},
		},
		{
			name:        "non binary file",
			inputFolder: "testdata/licensedir/withlicense",
			err:         nil,
			expected: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 0,
				NumberOfWarn: 0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)

			mockRepoClient.EXPECT().ListFiles(gomock.Any()).DoAndReturn(func(predicate func(string) (bool, error)) ([]string, error) {
				var files []string
				dirFiles, err := os.ReadDir(tt.inputFolder)
				if err == nil {
					for _, file := range dirFiles {
						files = append(files, file.Name())
					}
					print(files)
				}
				return files, err
			}).AnyTimes()

			mockRepoClient.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(func(file string) (io.ReadCloser, error) {
				return os.Open("./" + tt.inputFolder + "/" + file)
			}).AnyTimes()

			mockProjectClient := mockrepo.NewMockProjectPackageClient(ctrl)
			mockProjectClient.EXPECT().GetPackageName().DoAndReturn(func() string {
				return "fake-name"
			}).AnyTimes()

			mockProjectClient.EXPECT().GetSystem().DoAndReturn(func() string {
				return "fake-url"
			}).AnyTimes()

			mockProjectClient.EXPECT().GetPackageDependencies(gomock.Any()).DoAndReturn(func(ctx context.Context) (*packageclient.PackageDependencies, error) {
				return nil, fmt.Errorf("Error")
			}).AnyTimes()

			ctx := context.Background()

			dl := scut.TestDetailLogger{}

			req := checker.CheckRequest{
				Ctx:           ctx,
				RepoClient:    mockRepoClient,
				Dlogger:       &dl,
				ProjectClient: mockProjectClient,
			}

			result := BinaryArtifacts(&req)

			scut.ValidateTestReturn(t, tt.name, &tt.expected, &result, &dl)

			ctrl.Finish()
		})
	}
}

// currently only tests up to two dependencies. Each simulated dependency corresponds to a specified folder in inputFolders.
func TestBinaryArtifactsWithDependencies(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		inputFolders []string // one per each simulated dependency
		err          error
		expected     scut.TestReturn
	}{
		{
			name:         "one w/ jar files, one w/ no binary file",
			inputFolders: []string{"testdata/binaryartifacts/jars", "testdata/licensedir/withlicense"},
			err:          nil,
			expected: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 0,
				NumberOfWarn: 2,
			},
		},
		{
			name:         "only one w/ non binary file",
			inputFolders: []string{"testdata/licensedir/withlicense"},
			err:          nil,
			expected: scut.TestReturn{
				Score:        checker.MaxResultScore,
				NumberOfInfo: 0,
				NumberOfWarn: 0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			// mock RepoClient for first dependency
			firstMockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			firstMockRepoClient.EXPECT().ListFiles(gomock.Any()).DoAndReturn(func(predicate func(string) (bool, error)) ([]string, error) {
				var files []string
				dirFiles, err := os.ReadDir(tt.inputFolders[0])
				if err == nil {
					for _, file := range dirFiles {
						files = append(files, file.Name())
					}
					print(files)
				}
				return files, err
			}).AnyTimes()

			firstMockRepoClient.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(func(file string) (io.ReadCloser, error) {
				return os.Open("./" + tt.inputFolders[0] + "/" + file)
			}).AnyTimes()

			firstMockRepoClient.EXPECT().InitRepo(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(clients.Repo, string, int) error {
				return nil
			}).AnyTimes()

			// mock RepoClient for second dependency in test (if it exists)
			secondMockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			secondMockRepoClient.EXPECT().ListFiles(gomock.Any()).DoAndReturn(func(predicate func(string) (bool, error)) ([]string, error) {
				var files []string
				dirFiles, err := os.ReadDir(tt.inputFolders[1])
				if err == nil {
					for _, file := range dirFiles {
						files = append(files, file.Name())
					}
					print(files)
				}
				return files, err
			}).AnyTimes()

			secondMockRepoClient.EXPECT().GetFileReader(gomock.Any()).DoAndReturn(func(file string) (io.ReadCloser, error) {
				return os.Open("./" + tt.inputFolders[1] + "/" + file)
			}).AnyTimes()

			secondMockRepoClient.EXPECT().InitRepo(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(clients.Repo, string, int) error {
				return nil
			}).AnyTimes()

			// Mock package client
			mockPkgC := mockrepo.NewMockProjectPackageClient(ctrl)
			mockPkgC.EXPECT().GetPackageDependencies(gomock.Any()).DoAndReturn(
				func(ctx context.Context) (*packageclient.PackageDependencies, error) {
					v := packageclient.PackageDependencies{}

					// Add a simulated dependency for each item in inputFolders
					for range tt.inputFolders {
						v.Nodes = append(v.Nodes, struct {
							VersionKey struct {
								System  string "json:\"system\""
								Name    string "json:\"name\""
								Version string "json:\"version\""
							}
							Bundled  bool     "json:\"bundled\""
							Relation string   "json:\"relation\""
							Errors   []string "json:\"errors\""
						}{
							VersionKey: struct {
								System  string "json:\"system\""
								Name    string "json:\"name\""
								Version string "json:\"version\""
							}{
								System:  "GO",
								Name:    "Package",
								Version: "v0.1.0",
							},
						})
					}

					return &v, nil
				},
			).AnyTimes()

			mockPkgC.EXPECT().GetPackageName().DoAndReturn(
				func() string {
					return "name"
				},
			).AnyTimes()

			mockPkgC.EXPECT().GetSystem().DoAndReturn(
				func() string {
					return "system"
				},
			).AnyTimes()

			mockPkgC.EXPECT().GetURI(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(context.Context, string, string, string) (string, error) {
					return "github.com/ossf/scorecard", nil
				},
			).AnyTimes()

			firstCreate := mockPkgC.EXPECT().CreateGithubRepoClient(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, l *sclog.Logger) clients.RepoClient {
					return firstMockRepoClient
				},
			).MaxTimes(1)

			mockPkgC.EXPECT().CreateGithubRepoClient(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, l *sclog.Logger) clients.RepoClient {
					return secondMockRepoClient
				},
			).MaxTimes(1).After(firstCreate)

			parentMockRepoClient := mockrepo.NewMockRepoClient(ctrl)

			parentMockRepoClient.EXPECT().ListFiles(gomock.Any()).DoAndReturn(func(predicate func(string) (bool, error)) ([]string, error) {
				var files []string
				// dirFiles, err := os.ReadDir(tt.inputFolder)
				// if err == nil {
				// 	for _, file := range dirFiles {
				// 		files = append(files, file.Name())
				// 	}
				// 	print(files)
				// }
				return files, nil
			}).AnyTimes()

			ctx := context.Background()

			dl := scut.TestDetailLogger{}

			repo, _ := githubrepo.MakeGithubRepo("ossf/scorecard") // just to avoid panic. Actual value not critical
			// if err != nil {
			// 	t.Fatalf(`githubrepo.MakeGithubRepo() failed, error`, err)
			// }
			req := checker.CheckRequest{
				Ctx:           ctx,
				Dlogger:       &dl,
				ProjectClient: mockPkgC,
				Repo:          repo,
				RepoClient:    parentMockRepoClient,
			}

			result := BinaryArtifacts(&req)

			scut.ValidateTestReturn(t, tt.name, &tt.expected, &result, &dl)
			req.Dlogger.Flush()
			ctrl.Finish()
		})
	}
}
