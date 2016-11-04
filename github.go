// Copyright 2016 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"github.com/juju/errors"
	"github.com/ngaut/log"
	"golang.org/x/oauth2"
)

func newClient(token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)
	return client
}

func listPublicRepos(client *github.Client, org string) ([]*github.Repository, error) {
	opt := &github.RepositoryListByOrgOptions{Type: "public"}
	repos, _, err := client.Repositories.ListByOrg(org, opt)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return repos, nil
}

func printRepos(repos []*github.Repository) {
	var names []string
	for _, repo := range repos {
		names = append(names, *repo.Name)
	}

	sort.Strings(names)

	content := strings.Join(names, "\n")
	log.Infof("[repos]\n%s", content)
}

type UserSlice []*github.User

func (s UserSlice) Len() int           { return len(s) }
func (s UserSlice) Less(i, j int) bool { return *s[i].Login < *s[j].Login }
func (s UserSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func listStargazers(client *github.Client, owner string, repo string, onlyID bool) ([]*github.User, error) {
	opt := &github.ListOptions{PerPage: 100}

	var users []*github.User
	for {
		stargazers, resp, err := client.Activity.ListStargazers(owner, repo, opt)
		if err != nil {
			return nil, errors.Trace(err)
		}

		for _, stargazer := range stargazers {
			var user *github.User

			if onlyID {
				user = stargazer.User
			} else {
				user, _, err = client.Users.GetByID(*stargazer.User.ID)
				if err != nil {
					return nil, errors.Trace(err)
				}
			}

			users = append(users, user)
		}

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	if !onlyID {
		sort.Sort(UserSlice(users))
	}

	return users, nil
}

func listStargazersFromFile(client *github.Client, file string) ([]*github.User, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer f.Close()

	var users []*github.User

	br := bufio.NewReader(f)
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		} else {
			id, err := strconv.ParseInt(strings.TrimSpace(line), 10, 64)
			if err != nil {
				return nil, errors.Trace(err)
			}

			user, _, err := client.Users.GetByID(int(id))
			if err != nil {
				return nil, errors.Trace(err)
			}

			users = append(users, user)
		}
	}

	return users, nil
}

func printStargazers(owner string, repo string, users []*github.User, onlyID bool) {
	var content []byte
	for _, user := range users {
		if onlyID {
			content = append(content, []byte(unifyInt(user.ID))...)
			content = append(content, '\n')
		} else {
			content = append(content, []byte(fmt.Sprintf("%s/%s", owner, repo))...)
			content = append(content, '\t')
			content = append(content, []byte(unifyInt(user.ID))...)
			content = append(content, '\t')
			content = append(content, []byte(unifyStr(user.Login))...)
			content = append(content, '\t')
			content = append(content, []byte(unifyStr(user.Name))...)
			content = append(content, '\t')
			content = append(content, []byte(unifyStr(user.Email))...)
			content = append(content, '\t')
			content = append(content, []byte(unifyStr(user.Location))...)
			content = append(content, '\t')
			content = append(content, []byte(unifyStr(user.Company))...)
			content = append(content, '\t')
			content = append(content, []byte(unifyStr(user.Blog))...)
			content = append(content, '\t')
			content = append(content, []byte(unifyStr(user.Bio))...)
			content = append(content, '\t')
			content = append(content, []byte(unifyInt(user.PublicRepos))...)
			content = append(content, '\t')
			content = append(content, []byte(unifyInt(user.Following))...)
			content = append(content, '\t')
			content = append(content, []byte(unifyInt(user.Followers))...)
			content = append(content, '\t')
			content = append(content, []byte(unifyStr(user.HTMLURL))...)
			content = append(content, '\n')
		}
	}

	log.Infof("[stargazers][user]\n%s", string(content))
}

func unifyStr(s *string) string {
	if s == nil {
		return ""
	}

	ss := *s
	strings.Replace(ss, "\t", "  ", -1)
	return ss
}

func unifyInt(i *int) string {
	return fmt.Sprintf("%d", *i)
}
