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
	case "repos":
		repos, err := listPublicRepos(client, cfg.Owner)
		if err != nil {
			log.Fatal(err)
		}

		printRepos(repos)
	case "stargazers":
		if len(cfg.Input) > 0 {
			users, err := listStargazersFromFile(client, cfg.Input)
			if err != nil {
				log.Fatal(err)
			}

			printStargazers(cfg.Owner, cfg.Repo, users, false)
		} else {
			users, err := listStargazers(client, cfg.Owner, cfg.Repo, false)
			if err != nil {
				log.Fatal(err)
			}

			printStargazers(cfg.Owner, cfg.Repo, users, false)
		}
	case "stargazer-ids":
		users, err := listStargazers(client, cfg.Owner, cfg.Repo, true)
		if err != nil {
			log.Fatal(err)
		}

		printStargazers(cfg.Owner, cfg.Repo, users, true)
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
