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
	"fmt"
	"time"
)

// Return difference between two times in a printable format
func getTimeDiffString(a, b time.Time) (timeDiff string) {
	year, month, day, hour, min, sec := getTimeDiffNumbers(a, b)

	// Format the output

	timeDiff = ""
	if year > 0 {
		if year > 1 {
			timeDiff += fmt.Sprintf("%d years, ", year)
		} else {
			timeDiff += fmt.Sprintf("%d year, ", year)
		}
	}

	if len(timeDiff) > 0 || month > 0 {
		if month == 0 || month > 1 {
			timeDiff += fmt.Sprintf("%d months, ", month)
		} else {
			timeDiff += fmt.Sprintf("%d month, ", month)
		}
	}

	if len(timeDiff) > 0 || day > 0 {
		if day == 0 || day > 1 {
			timeDiff += fmt.Sprintf("%d days, ", day)
		} else {
			timeDiff += fmt.Sprintf("%d day, ", day)
		}
	}

	switch {
	case len(timeDiff) > 0 || hour > 0:
		timeDiff += fmt.Sprintf("%d:%2.2d:%2.2d", hour, min, sec)
	case min > 0:
		timeDiff = fmt.Sprintf("%d:%2.2d", min, sec)
	default:
		if sec == 0 || sec > 1 {
			timeDiff = fmt.Sprintf("%d seconds", sec)
		} else {
			timeDiff = fmt.Sprintf("%d second", sec)
		}
	}

	return
}

func daysIn(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// Return difference between two times numerically
// Source: https://stackoverflow.com/questions/36530251/golang-time-since-with-months-and-years
func getTimeDiffNumbers(from, to time.Time) (years, months, days, hours, minutes, seconds int) {
	if from.Location() != to.Location() {
		to = to.In(to.Location())
	}

	if from.After(to) {
		from, to = to, from
	}

	y1, M1, d1 := from.Date()
	y2, M2, d2 := to.Date()

	h1, m1, s1 := from.Clock()
	h2, m2, s2 := to.Clock()

	years = y2 - y1
	months = int(M2 - M1)
	days = d2 - d1

	hours = h2 - h1
	minutes = m2 - m1
	seconds = s2 - s1

	if seconds < 0 {
		seconds += 60
		minutes--
	}
	if minutes < 0 {
		minutes += 60
		hours--
	}
	if hours < 0 {
		hours += 24
		days--
	}
	if days < 0 {
		days += daysIn(y2, M2-1)
		months--
	}
	if months < 0 {
		months += 12
		years--
	}

	return
}
