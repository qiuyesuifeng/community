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
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/juju/errors"
)

// NewConfig creates a new config.
func NewConfig() *Config {
	cfg := &Config{}
	cfg.FlagSet = flag.NewFlagSet("community", flag.ContinueOnError)
	fs := cfg.FlagSet

	fs.StringVar(&cfg.configFile, "config", "", "Config file")

	fs.StringVar(&cfg.Owner, "o", "", "github owner")
	fs.StringVar(&cfg.Repo, "r", "", "github repo")
	fs.StringVar(&cfg.Token, "t", "", "github access token")

	fs.StringVar(&cfg.Service, "s", "", "print service information [ forkers | repos | stargazers | stargazer-ids | users | watchers ]")

	fs.StringVar(&cfg.Input, "i", "", "input file")

	fs.StringVar(&cfg.StartDate, "start", "", "start date")
	fs.StringVar(&cfg.EndDate, "end", "", "end date")

	fs.StringVar(&cfg.LogLevel, "L", "info", "log level: debug, info, warn, error, fatal")

	return cfg
}

// Config is the configuration.
type Config struct {
	*flag.FlagSet `json:"-"`

	Owner string `toml:"owner" json:"owner"`

	Repo string `toml:"repo" json:"repo"`

	Token string `toml:"token" json:"token"`

	Service string `json:"-"`

	StartDate string `json:"-"`

	EndDate string `json:"-"`

	Input string `json:"-"`

	LogLevel string `toml:"log-level" json:"log-level"`

	configFile string
}

// Parse parses flag definitions from the argument list.
func (c *Config) Parse(arguments []string) error {
	// Parse first to get config file.
	err := c.FlagSet.Parse(arguments)
	if err != nil {
		return errors.Trace(err)
	}

	// Load config file if specified.
	if c.configFile != "" {
		err = c.configFromFile(c.configFile)
		if err != nil {
			return errors.Trace(err)
		}
	}

	// Parse again to replace with command line options.
	err = c.FlagSet.Parse(arguments)
	if err != nil {
		return errors.Trace(err)
	}

	if len(c.FlagSet.Args()) != 0 {
		return errors.Errorf("'%s' is an invalid flag", c.FlagSet.Arg(0))
	}

	return nil
}

func (c *Config) String() string {
	if c == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Config(%+v)", *c)
}

// configFromFile loads config from file.
func (c *Config) configFromFile(path string) error {
	_, err := toml.DecodeFile(path, c)
	return errors.Trace(err)
}
