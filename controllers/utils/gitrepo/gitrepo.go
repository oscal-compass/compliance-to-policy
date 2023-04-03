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
	"time"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

type GitRepo struct {
	repo        *git.Repository
	dir         string
	username    string
	accessToken string
}

func NewGitRepo(dir string, cloneOpts git.CloneOptions) (GitRepo, error) {
	var gitRepo GitRepo
	dir, err := os.MkdirTemp(dir, "tmp-")
	if err != nil {
		return gitRepo, err
	}
	repo, err := git.PlainClone(dir, false, &cloneOpts)
	if err != nil {
		return gitRepo, err
	}
	gitRepo = GitRepo{repo: repo, dir: dir}
	return gitRepo, nil
}

func NewGitRepoWithAuth(dir string, url string, username string, accessToken string) (GitRepo, error) {
	opts := git.CloneOptions{
		URL:  url,
		Auth: &http.BasicAuth{Username: username, Password: accessToken},
	}
	gitRepo, err := NewGitRepo(dir, opts)
	if err != nil {
		return gitRepo, err
	}
	gitRepo.username = username
	gitRepo.accessToken = accessToken
	return gitRepo, nil
}

func (g *GitRepo) GetDirectory() string {
	return g.dir
}

func (g *GitRepo) Checkout(branch string) error {
	n := plumbing.NewBranchReferenceName("refs/heads/" + branch)
	w, err := g.repo.Worktree()
	if err != nil {
		return err
	}
	return w.Checkout(&git.CheckoutOptions{
		Branch: n,
		Create: false,
	})
}

func (g *GitRepo) Commit(path string, message string) error {
	w, err := g.repo.Worktree()
	if err != nil {
		return err
	}
	_, err = w.Add(path)
	if err != nil {
		return err
	}
	_, err = w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "application",
			Email: "application@local",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (g *GitRepo) Push() error {
	return g.repo.Push(&git.PushOptions{
		Auth: &http.BasicAuth{Username: g.username, Password: g.accessToken},
	})
}
