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
	"fmt"
	"sort"
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

type StargazerSlice []*github.Stargazer

func (s StargazerSlice) Len() int           { return len(s) }
func (s StargazerSlice) Less(i, j int) bool { return *s[i].User.Login < *s[j].User.Login }
func (s StargazerSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func listStargazers(client *github.Client, owner string, repo string) ([]*github.Stargazer, error) {
	opt := &github.ListOptions{PerPage: 100}

	var allStargazers []*github.Stargazer
	for {
		stargazers, resp, err := client.Activity.ListStargazers(owner, repo, opt)
		if err != nil {
			return nil, errors.Trace(err)
		}

		for _, stargazer := range stargazers {
			user, _, err := client.Users.GetByID(*stargazer.User.ID)
			if err != nil {
				return nil, errors.Trace(err)
			}

			stargazer.User = user
			allStargazers = append(allStargazers, stargazer)
		}

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	sort.Sort(StargazerSlice(allStargazers))
	return allStargazers, nil
}

func printStargazers(owner string, repo string, stargazers []*github.Stargazer) {
	var content []byte
	for _, stargazer := range stargazers {
		content = append(content, []byte(fmt.Sprintf("%s/%s", owner, repo))...)
		content = append(content, '\t')
		content = append(content, []byte(unifyInt(stargazer.User.ID))...)
		content = append(content, '\t')
		content = append(content, []byte(unifyStr(stargazer.User.Login))...)
		content = append(content, '\t')
		content = append(content, []byte(unifyStr(stargazer.User.Name))...)
		content = append(content, '\t')
		content = append(content, []byte(unifyStr(stargazer.User.Email))...)
		content = append(content, '\t')
		content = append(content, []byte(unifyStr(stargazer.User.Location))...)
		content = append(content, '\t')
		content = append(content, []byte(unifyStr(stargazer.User.Company))...)
		content = append(content, '\t')
		content = append(content, []byte(unifyStr(stargazer.User.Blog))...)
		content = append(content, '\t')
		content = append(content, []byte(unifyStr(stargazer.User.Bio))...)
		content = append(content, '\t')
		content = append(content, []byte(unifyInt(stargazer.User.PublicRepos))...)
		content = append(content, '\t')
		content = append(content, []byte(unifyInt(stargazer.User.Following))...)
		content = append(content, '\t')
		content = append(content, []byte(unifyInt(stargazer.User.Followers))...)
		content = append(content, '\t')
		content = append(content, []byte(unifyStr(stargazer.User.HTMLURL))...)
		content = append(content, '\n')
	}

	if len(content) > 0 {
		content = content[:len(content)-1]
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
