// Copyright 2023 OpenSSF Scorecard Authors
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

package scorecard_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//nolint:paralleltest // avoiding parallel e2e tests due to rate limit concerns (#2527)
func TestPkg(t *testing.T) {
	if val, exists := os.LookupEnv("SKIP_GINKGO"); exists && val == "1" {
		t.Skip()
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pkg Suite")
}
