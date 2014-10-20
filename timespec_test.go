package timespec

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"
)

type testTime struct {
	input    string
	expected *Time
}

func TestParseTime(t *testing.T) {
	for _, testcase := range []testTime{
		{"1 pm", &Time{hours: 13}},
		{"12 pm", &Time{hours: 12}},
		{"11 pm", &Time{hours: 23}},
		{"11:59 pm", &Time{hours: 23, minutes: 59}},
		{"12:10 UTC", &Time{hours: 12, minutes: 10, timezoneName: "UTC"}},
		{"12:10 utc", &Time{hours: 12, minutes: 10, timezoneName: "UTC"}},
		{"13 UTC", &Time{hours: 13, timezoneName: "UTC"}},
		{"1 am", &Time{hours: 1}},
		{"13:15", &Time{hours: 13, minutes: 15}},
		{"12 uTC", &Time{hours: 12, timezoneName: "UTC"}},
		{"1215", &Time{hours: 12, minutes: 15}},
		{"0512 utC", &Time{hours: 5, minutes: 12, timezoneName: "UTC"}},
		{"noon", &Time{hours: 12}},
		{"midnight", &Time{}},
	} {
		src := bufio.NewReader(bytes.NewBufferString(testcase.input))
		result := Time{}
		err := parseTime(src, &result)

		if err != nil {
			t.Fatalf("parseTime(%q): %s", testcase.input, err)
		}

		if !reflect.DeepEqual(&result, testcase.expected) {
			t.Fatalf("parseTime(%q):\n  Expected: %#v\n       Got: %#v\n",
				testcase.input, testcase.expected, &result)
		}
	}
}

type testDate struct {
	input    string
	expected *Date
}

func TestParseDate(t *testing.T) {
	for _, testcase := range []*testDate{
		{"Feb 02", &Date{month: 2, day: 2}},
		{"Mar 11, 2010", &Date{month: 3, day: 11, year: 2010}},
		{"tomorrow", &Date{isTomorrow: true}},
		{"today", &Date{isToday: true}},
		{"December 24 , 2015", &Date{month: 12, day: 24, year: 2015}},
	} {
		src := bufio.NewReader(bytes.NewBufferString(testcase.input))
		result := Date{}
		err := parseDate(src, &result)

		if err != nil {
			t.Fatalf("parseDate(%q): %s", testcase.input, err)
		}

		if !reflect.DeepEqual(&result, testcase.expected) {
			t.Fatalf("parseDate(%q):\n  Expected: %#v\n       Got: %#v\n",
				testcase.input, testcase.expected, &result)
		}

	}
}

type testIncrement struct {
	input    string
	expected *Increment
}

func TestParseIncrement(t *testing.T) {
	for _, testcase := range []*testIncrement{
		{"+1 day", &Increment{1, IncrementDays}},
		{"+ 1 day", &Increment{1, IncrementDays}},
		{"next week", &Increment{1, IncrementWeeks}},
		{"nextday", &Increment{1, IncrementDays}},
		{"+ 20 months", &Increment{20, IncrementMonths}},
	} {
		src := bufio.NewReader(bytes.NewBufferString(testcase.input))
		result := Increment{}
		err := parseIncrement(src, &result)

		if err != nil {
			t.Fatalf("parseIncrement(%q): %s", testcase.input, err)
		}

		if !reflect.DeepEqual(&result, testcase.expected) {
			t.Fatalf("parseIncrement(%q):\n  Expected: %#v\n       Got: %#v\n",
				testcase.input, testcase.expected, &result)
		}

	}
}

type timespecTest struct {
	input    string
	expected *Timespec
}

func TestParseTimespec(t *testing.T) {
	for _, testcase := range []*timespecTest{
		{"now + 1 day", &Timespec{
			increment: &Increment{1, IncrementDays},
			time:      &Time{isNow: true},
			date:      &Date{},
		}},
		{"now", &Timespec{
			increment: &Increment{},
			time:      &Time{isNow: true},
			date:      &Date{},
		}},
		{"12:11", &Timespec{
			increment: &Increment{},
			time:      &Time{hours: 12, minutes: 11},
			date:      &Date{isToday: true},
		}},
		{"10 am next week", &Timespec{
			increment: &Increment{1, IncrementWeeks},
			time:      &Time{hours: 10},
			date:      &Date{isToday: true},
		}},
		{"14:00 Feb 12, 2015 + 3 week", &Timespec{
			increment: &Increment{3, IncrementWeeks},
			time:      &Time{hours: 14},
			date:      &Date{month: 2, day: 12, year: 2015},
		}},
		{"9:00 UTCnextweek", &Timespec{
			increment: &Increment{1, IncrementWeeks},
			time:      &Time{hours: 9, timezoneName: "UTC"},
			date:      &Date{isToday: true},
		}},
	} {
		src := bufio.NewReader(bytes.NewBufferString(testcase.input))
		result := Timespec{
			increment: &Increment{},
			time:      &Time{},
			date:      &Date{},
		}
		err := parseTimespec(src, &result)

		if err != nil {
			t.Fatalf("parseTimespec(%q): %s", testcase.input, err)
		}

		if !reflect.DeepEqual(&result, testcase.expected) {
			t.Fatalf(`parseTimespec(%q):
Expected:
time: %#v,
date: %#v,
incr: %#v

Got:
time: %#v,
date: %#v,
incr: %#v
`, testcase.input, testcase.expected.time, testcase.expected.date, testcase.expected.increment, result.time, result.date, result.increment)
		}
	}
}
