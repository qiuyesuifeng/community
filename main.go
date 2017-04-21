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
	"flag"
	"os"
	"strings"

	"github.com/juju/errors"
	"github.com/ngaut/log"
)

func do(cfg *Config) {
	client := newClient(cfg.Token)

	switch strings.ToLower(cfg.Service) {
	case "contributors":
		if len(cfg.Owner) == 0 {
			log.Fatal("empty owner")
		}

		if len(cfg.Repo) == 0 {
			log.Fatal("empty repo")
		}

		users, times, err := listCommits(client, cfg)
		if err != nil {
			log.Fatal(err)
		}

		printUserNames(cfg.Owner, cfg.Repo, users, times)
	case "forkers":
		if len(cfg.Owner) == 0 {
			log.Fatal("empty owner")
		}

		if len(cfg.Repo) == 0 {
			log.Fatal("empty repo")
		}

		users, times, err := listForkers(client, cfg)
		if err != nil {
			log.Fatal(err)
		}

		printUsers(cfg.Owner, cfg.Repo, users, times)
	case "issues":
		if len(cfg.Owner) == 0 {
			log.Fatal("empty owner")
		}

		if len(cfg.Repo) == 0 {
			log.Fatal("empty repo")
		}

		users, err := listIssues(client, cfg)
		if err != nil {
			log.Fatal(err)
		}

		printUsers(cfg.Owner, cfg.Repo, users, nil)
	case "repos":
		if len(cfg.Owner) == 0 {
			log.Fatal("empty owner")
		}

		repos, err := listPublicRepos(client, cfg.Owner)
		if err != nil {
			log.Fatal(err)
		}

		printRepos(repos)
	case "stargazers":
		if len(cfg.Owner) == 0 {
			log.Fatal("empty owner")
		}

		if len(cfg.Repo) == 0 {
			log.Fatal("empty repo")
		}

		users, times, err := listStargazers(client, cfg, false)
		if err != nil {
			log.Fatal(err)
		}

		printUsers(cfg.Owner, cfg.Repo, users, times)
	case "stargazer-ids":
		if len(cfg.Owner) == 0 {
			log.Fatal("empty owner")
		}

		if len(cfg.Repo) == 0 {
			log.Fatal("empty repo")
		}

		users, times, err := listStargazers(client, cfg, true)
		if err != nil {
			log.Fatal(err)
		}

		printUserIDs(users, times)
	case "users":
		if len(cfg.Input) == 0 {
			log.Fatal("empty input")
		}

		users, err := listUsers(client, cfg.Input)
		if err != nil {
			log.Fatal(err)
		}

		printUsers("", "", users, nil)
	case "watchers":
		if len(cfg.Owner) == 0 {
			log.Fatal("empty owner")
		}

		if len(cfg.Repo) == 0 {
			log.Fatal("empty repo")
		}

		users, _, err := listWatchers(client, cfg)
		if err != nil {
			log.Fatal(err)
		}

		printUsers(cfg.Owner, cfg.Repo, users, nil)
	}
}

func main() {
	cfg := NewConfig()
	err := cfg.Parse(os.Args[1:])
	switch errors.Cause(err) {
	case nil:
	case flag.ErrHelp:
		os.Exit(0)
	default:
		log.Errorf("parse cmd flags err - %s\n", err)
		os.Exit(2)
	}

	do(cfg)
}
