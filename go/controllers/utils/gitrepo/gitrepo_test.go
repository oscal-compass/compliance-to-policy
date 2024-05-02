/*
Copyright 2023 IBM Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gitrepo

import (
	"os"
	"testing"

	"github.com/oscal-compass/compliance-to-policy/go/pkg"
)

func TestGitRepo(t *testing.T) {
	if _, ok := os.LookupEnv("DO_NOT_SKIP_ANY_TEST"); !ok {
		t.Skip("Skipping testing")
	}
	dir, err := pkg.MakeDir(pkg.PathFromPkgDirectory("../controllers/utils/gitrepo/_test"))
	if err != nil {
		panic(err)
	}
	username := os.Getenv("username")
	token := os.Getenv("token")
	url := os.Getenv("url")
	gitRepo, err := NewGitRepoWithAuth(dir, url, username, token)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(gitRepo.GetDirectory()+"/test1.txt", []byte("hoge"), os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.WriteFile(gitRepo.GetDirectory()+"/test2.txt", []byte("foo"), os.ModePerm); err != nil {
		panic(err)
	}
	if err := gitRepo.Commit(".", "test"); err != nil {
		panic(err)
	}
	if err := gitRepo.Push(); err != nil {
		panic(err)
	}
}
