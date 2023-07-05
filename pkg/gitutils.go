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

package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

type GitUtils struct {
	gitRepoCache map[string]string
	tempDir      TempDirectory
}

func NewGitUtils(tempDir TempDirectory) GitUtils {
	return GitUtils{
		gitRepoCache: map[string]string{},
		tempDir:      tempDir,
	}
}

func (g *GitUtils) LoadFromWeb(url string, out interface{}) error {
	u, err := neturl.Parse(url)
	if err != nil {
		return err
	}
	if u.Scheme == "local" {
		return loadFromLocalFs(u, out)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("Failed to initialize http client for %s", url)
	}

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to get %s", url)
	}
	defer resp.Body.Close()

	byteArray, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to serialize body %s", url)
	}

	err = json.Unmarshal(byteArray, out)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal %s", url)
	}
	return nil
}

func (g *GitUtils) LoadFromGit(url string, out interface{}) error {
	u, err := neturl.Parse(url)
	if err != nil {
		return err
	}
	paths := strings.Split(u.Path, "/")
	repoUrl := fmt.Sprintf("%s://%s/%s/%s", u.Scheme, u.Host, paths[1], paths[2])
	path := strings.Join(paths[3:], "/")

	if u.Scheme == "local" {
		return loadFromLocalFs(u, out)
	}
	repoDir, _, err := g.GitClone(repoUrl)
	if err != nil {
		return fmt.Errorf("Failed to clone %s", repoUrl)
	}
	if err := LoadJsonFileToObject(repoDir+"/"+path, out); err != nil {
		return fmt.Errorf("Failed to marshal %s", repoDir+path)
	}
	return nil
}

func loadFromLocalFs(u *neturl.URL, out interface{}) error {
	path := toLocalPath(u)
	if err := LoadJsonFileToObject(path, out); err != nil {
		return fmt.Errorf("Failed to marshal %s in local directory", path)
	}
	return nil
}

func toLocalPath(u *neturl.URL) string {
	return u.Host + u.Path
}

func (g *GitUtils) GitClone(url string) (string, string, error) {
	u, err := neturl.Parse(url)
	if err != nil {
		return "", "", err
	}
	paths := strings.Split(u.Path, "/")
	repoUrl := fmt.Sprintf("%s://%s/%s/%s", u.Scheme, u.Host, paths[1], paths[2])
	path := strings.Join(paths[3:], "/")
	if u.Scheme == "local" {
		return toLocalPath(u), "", nil
	}
	rootDir, err := g.gitClone(repoUrl)
	return rootDir, path, err
}

func (g *GitUtils) gitClone(url string) (string, error) {
	username := os.Getenv("username")
	token := os.Getenv("token")
	dir, ok := g.gitRepoCache[url]
	if !ok {
		dir, err := os.MkdirTemp(g.tempDir.GetTempDir(), "tmp-")
		if err != nil {
			return "", err
		}
		cloneOption := &git.CloneOptions{
			URL: url,
		}
		if username != "" && token != "" {
			logger.Info("Git Clone with Auth given by 'username' and 'token' in environment variables ")
			cloneOption.Auth = &githttp.BasicAuth{Username: username, Password: token}
		}
		if _, err := git.PlainClone(dir, false, cloneOption); err != nil {
			return "", err
		}
		g.gitRepoCache[url] = dir
		return dir, nil
	}
	return dir, nil
}
