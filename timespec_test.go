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
		{"12:10 Europe/Berlin", &Time{hours: 12, minutes: 10, timezoneName: "Europe/Berlin"}},
		{"13 UTC", &Time{hours: 13, timezoneName: "UTC"}},
		{"1 am", &Time{hours: 1}},
		{"13:15", &Time{hours: 13, minutes: 15}},
		{"12 Europe/Berlin", &Time{hours: 12, timezoneName: "Europe/Berlin"}},
		{"1215", &Time{hours: 12, minutes: 15}},
		{"0512 UTC", &Time{hours: 5, minutes: 12, timezoneName: "UTC"}},
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