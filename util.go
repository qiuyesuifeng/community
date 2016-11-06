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
	"strings"
	"time"

	"github.com/juju/errors"
)

var (
	zeroTime time.Time
)

const (
	timeFormat = "2006-01-02"
)

func unifyStr(s *string) string {
	if s == nil {
		return ""
	}

	ss := *s
	ss = strings.Replace(ss, "\t", " ", -1)
	ss = strings.Replace(ss, "\n", " ", -1)
	ss = strings.Replace(ss, "\r", " ", -1)
	return ss
}

func unifyInt(i *int) string {
	return fmt.Sprintf("%d", *i)
}

func unifyTime(date string) (time.Time, error) {
	t, err := time.Parse(timeFormat, date)
	if err != nil {
		return zeroTime, errors.Trace(err)
	}

	return t, nil
}

func checkTime(start time.Time, end time.Time, t time.Time) bool {
	return start.Before(t) && end.After(t)
}
