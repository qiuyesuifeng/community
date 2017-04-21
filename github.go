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
	"time"

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
		log.Infof("[repo]%v", repo)
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

func listCommits(client *github.Client, cfg *Config) ([]string, []string, error) {
	opt := &github.CommitsListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var (
		users = make(map[string][]string)
	)
	for {
		commits, resp, err := client.Repositories.ListCommits(cfg.Owner, cfg.Repo, opt)
		if err != nil {
			return nil, nil, errors.Trace(err)
		}

		for _, commit := range commits {
			user := *commit.Commit.Author.Name
			t := unifyDate(*commit.Commit.Author.Date)
			value, ok := users[user]
			if ok {
				users[user] = append(value, t)
			} else {
				users[user] = []string{t}
			}
		}

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	var (
		userNames []string
		times     []string
	)
	for name, value := range users {
		userNames = append(userNames, name)
		sort.Strings(value)
		times = append(times, value[0])
	}

	return userNames, times, nil
}

func listForkers(client *github.Client, cfg *Config) ([]*github.User, []time.Time, error) {
	useTimeFilter := len(cfg.StartDate) > 0 && len(cfg.EndDate) > 0

	var (
		start time.Time
		end   time.Time
		err   error
	)

	if useTimeFilter {
		start, err = parseDate(cfg.StartDate)
		if err != nil {
			return nil, nil, errors.Trace(err)
		}

		end, err = parseDate(cfg.EndDate)
		if err != nil {
			return nil, nil, errors.Trace(err)
		}
	}

	opt := &github.RepositoryListForksOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var (
		users []*github.User
		times []time.Time
	)
	for {
		repos, resp, err := client.Repositories.ListForks(cfg.Owner, cfg.Repo, opt)
		if err != nil {
			return nil, nil, errors.Trace(err)
		}

		for _, repo := range repos {
			if useTimeFilter {
				if !checkTime(start, end, repo.CreatedAt.Time) {
					continue
				}
			}

			user, _, err := client.Users.GetByID(*repo.Owner.ID)
			if err != nil {
				return nil, nil, errors.Trace(err)
			}

			users = append(users, user)
			times = append(times, repo.CreatedAt.Time)
		}

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return users, times, nil
}

func listWatchers(client *github.Client, cfg *Config) ([]*github.User, []time.Time, error) {
	opt := &github.ListOptions{PerPage: 100}

	var (
		allUsers []*github.User
		times    []time.Time
	)
	for {
		users, resp, err := client.Activity.ListWatchers(cfg.Owner, cfg.Repo, opt)
		if err != nil {
			return nil, nil, errors.Trace(err)
		}

		for _, user := range users {
			user, _, err := client.Users.GetByID(*user.ID)
			if err != nil {
				return nil, nil, errors.Trace(err)
			}

			allUsers = append(allUsers, user)
			// TODO: add watch time
		}

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return allUsers, times, nil
}

func listIssues(client *github.Client, cfg *Config) ([]*github.User, error) {
	opt := &github.IssueListByRepoOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var users []*github.User
	userCache := make(map[int]struct{})
	for {
		issues, resp, err := client.Issues.ListByRepo(cfg.Owner, cfg.Repo, opt)
		if err != nil {
			return nil, errors.Trace(err)
		}

		for _, issue := range issues {
			_, ok := userCache[*issue.User.ID]
			if ok {
				continue
			}

			user, _, err := client.Users.GetByID(*issue.User.ID)
			if err != nil {
				return nil, errors.Trace(err)
			}

			users = append(users, user)
			userCache[*issue.User.ID] = struct{}{}
		}

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return users, nil
}

func listStargazers(client *github.Client, cfg *Config, onlyID bool) ([]*github.User, []time.Time, error) {
	opt := &github.ListOptions{PerPage: 100}
	useTimeFilter := len(cfg.StartDate) > 0 && len(cfg.EndDate) > 0

	var (
		start time.Time
		end   time.Time
		err   error
	)

	if useTimeFilter {
		start, err = parseDate(cfg.StartDate)
		if err != nil {
			return nil, nil, errors.Trace(err)
		}

		end, err = parseDate(cfg.EndDate)
		if err != nil {
			return nil, nil, errors.Trace(err)
		}
	}

	var (
		users []*github.User
		times []time.Time
	)
	for {
		stargazers, resp, err := client.Activity.ListStargazers(cfg.Owner, cfg.Repo, opt)
		if err != nil {
			return nil, nil, errors.Trace(err)
		}

		for _, stargazer := range stargazers {
			if useTimeFilter {
				if !checkTime(start, end, stargazer.StarredAt.Time) {
					continue
				}
			}

			var user *github.User

			if onlyID {
				user = stargazer.User
			} else {
				user, _, err = client.Users.GetByID(*stargazer.User.ID)
				if err != nil {
					return nil, nil, errors.Trace(err)
				}
			}

			users = append(users, user)
			times = append(times, stargazer.StarredAt.Time)
		}

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return users, times, nil
}

func listUsers(client *github.Client, file string) ([]*github.User, error) {
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
			datas := strings.Fields(strings.TrimSpace(line))

			id, err := strconv.ParseInt(datas[0], 10, 64)
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

func printUsers(owner string, repo string, users []*github.User, times []time.Time) {
	printTime := len(times) > 0

	var content []byte
	for i, user := range users {
		if len(owner) > 0 && len(repo) > 0 {
			content = append(content, []byte(fmt.Sprintf("%s/%s", owner, repo))...)
			content = append(content, '\t')
		}

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
		if printTime {
			content = append(content, '\t')
			content = append(content, []byte(unifyDate(times[i]))...)
		}
		content = append(content, '\n')
	}

	log.Infof("[users]\n%s", string(content))
}

func printUserNames(owner string, repo string, users []string, times []string) {
	var content []byte
	for i, user := range users {
		if len(owner) > 0 && len(repo) > 0 {
			content = append(content, []byte(fmt.Sprintf("%s/%s", owner, repo))...)
			content = append(content, '\t')
		}

		content = append(content, []byte(unifyStr(&user))...)
		content = append(content, '\t')
		content = append(content, []byte(times[i])...)
		content = append(content, '\n')
	}

	log.Infof("[user names]\n%s", string(content))
}

func printUserIDs(users []*github.User, times []time.Time) {
	printTime := len(times) > 0

	var content []byte
	for i, user := range users {
		content = append(content, []byte(unifyInt(user.ID))...)
		if printTime {
			content = append(content, '\t')
			content = append(content, []byte(unifyDate(times[i]))...)
		}
		content = append(content, '\n')
	}

	log.Infof("[user ids]\n%s", string(content))
}
