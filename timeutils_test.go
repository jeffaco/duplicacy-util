// Copyright © 2018 Jeff Coffler <jeff@taltos.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"time"
	"testing"
)

func TestTimeDiffNumbers(t *testing.T) {
	tables := []struct {
		time1  time.Time
		time2  time.Time
		year   int
		month  int
		day    int
		hour   int
		min    int
		second int
	}{
		{
			time.Date(2015, 5, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2016, 6, 2, 1, 1, 1, 1, time.UTC),
			1, 1, 1, 1, 1, 1,
		},
		{
			time.Date(2016, 1, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2016, 2, 1, 0, 0, 0, 0, time.UTC),
			0, 0, 30, 0, 0, 0,
		},
		{
			time.Date(2016, 2, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2016, 3, 1, 0, 0, 0, 0, time.UTC),
			0, 0, 28, 0, 0, 0,
		},
		{
			time.Date(2015, 2, 11, 0, 0, 0, 0, time.UTC),
			time.Date(2016, 1, 12, 0, 0, 0, 0, time.UTC),
			0, 11, 1, 0, 0, 0,
		},
		{
			time.Date(2015, 1, 11, 0, 0, 0, 0, time.UTC),
			time.Date(2015, 3, 10, 0, 0, 0, 0, time.UTC),
			0, 1, 27, 0, 0, 0,
		},
	}

	for _, table := range tables {
		year, month, day, hour, min, second := getTimeDiffNumbers(table.time1, table.time2)
		if year != table.year { t.Errorf("Year was incorrect, got %d, expected %d.", year, table.year) }
		if month != table.month { t.Errorf("Month was incorrect, got %d, expected %d.", month, table.month) }
		if day != table.day { t.Errorf("Day was incorrect, got %d, expected %d.", day, table.day) }
		if hour != table.hour { t.Errorf("Hour was incorrect, got %d, expected %d.", hour, table.hour) }
		if min != table.min { t.Errorf("Minute was incorrect, got %d, expected %d.", min, table.min) }
		if second != table.second { t.Errorf("Second was incorrect, got %d, expected %d.", second, table.second) }
	}
}

func TestTimeDiffString(t *testing.T) {
	tables := []struct {
		time1 time.Time
		time2 time.Time
		result string
	}{
		{
			time.Date(2015, 5, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2016, 6, 2, 1, 1, 1, 1, time.UTC),
			"1 year, 1 month, 1 day, 1:01:01",
		},
		{
			time.Date(2016, 1, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2016, 2, 1, 0, 0, 0, 0, time.UTC),
			"30 days, 0:00:00",
		},
		{
			time.Date(2016, 2, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2016, 3, 1, 0, 0, 0, 0, time.UTC),
			"28 days, 0:00:00",
		},
		{
			time.Date(2015, 2, 11, 0, 0, 0, 0, time.UTC),
			time.Date(2016, 1, 12, 0, 0, 0, 0, time.UTC),
			"11 months, 1 day, 0:00:00",
		},
		{
			time.Date(2015, 1, 11, 0, 0, 0, 0, time.UTC),
			time.Date(2015, 3, 10, 0, 0, 0, 0, time.UTC),
			"1 month, 27 days, 0:00:00",
		},
	}

	for _, table := range tables {
		result := getTimeDiffString(table.time1, table.time2)
		if result != table.result { t.Errorf("Result was incorrect, got '%s', expected '%s'.", result, table.result) }
	}
}
