package timespec

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"
)

type testTimespec struct {
	input    string
	expected *Timespec
}

func TestParseTime(t *testing.T) {
	for _, testcase := range []testTimespec{
		{"1 pm", &Timespec{hours: 13}},
		{"12 pm", &Timespec{hours: 12}},
		{"11 pm", &Timespec{hours: 23}},
		{"11:59 pm", &Timespec{hours: 23, minutes: 59}},
		{"12:10 UTC", &Timespec{hours: 12, minutes: 10}},
		{"12:10 utc", &Timespec{hours: 12, minutes: 10}},
		{"13 UTC", &Timespec{hours: 13}},
		{"1 am", &Timespec{hours: 1}},
		{"13:15", &Timespec{hours: 13, minutes: 15}},
		{"12 uTC", &Timespec{hours: 12}},
		{"1215", &Timespec{hours: 12, minutes: 15}},
		{"0512 utC", &Timespec{hours: 5, minutes: 12}},
		{"noon", &Timespec{hours: 12}},
		{"midnight", &Timespec{}},
	} {
		src := bufio.NewReader(bytes.NewBufferString(testcase.input))
		result := Timespec{}
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

func TestParseDate(t *testing.T) {
	for _, testcase := range []*testTimespec{
		{"Feb 02", &Timespec{month: 2, day: 2}},
		{"Mar 11, 2010", &Timespec{month: 3, day: 11, year: 2010}},
		{"tomorrow", &Timespec{isTomorrow: true}},
		{"today", &Timespec{}},
		{"December 24 , 2015", &Timespec{month: 12, day: 24, year: 2015}},
	} {
		src := bufio.NewReader(bytes.NewBufferString(testcase.input))
		result := Timespec{}
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

func TestParseIncrement(t *testing.T) {
	for _, testcase := range []*testTimespec{
		{"+1 day", &Timespec{increments: 1, unit: IncrementDays}},
		{"+ 1 day", &Timespec{increments: 1, unit: IncrementDays}},
		{"next week", &Timespec{increments: 1, unit: IncrementWeeks}},
		{"nextday", &Timespec{increments: 1, unit: IncrementDays}},
		{"+ 20 months", &Timespec{increments: 20, unit: IncrementMonths}},
	} {
		src := bufio.NewReader(bytes.NewBufferString(testcase.input))
		result := Timespec{}
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

func TestParseTimespec(t *testing.T) {
	for _, testcase := range []*testTimespec{
		{"now + 1 day", &Timespec{
			increments: 1,
			unit:       IncrementDays,
			isNow:      true,
		}},
		{"now", &Timespec{isNow: true}},
		{"12:11", &Timespec{
			hours:   12,
			minutes: 11,
		}},
		{"10 am next week", &Timespec{
			increments: 1,
			unit:       IncrementWeeks,
			hours:      10,
		}},
		{"14:00 Feb 12, 2015 + 3 week", &Timespec{
			increments: 3,
			unit:       IncrementWeeks,
			hours:      14,
			month:      2,
			day:        12,
			year:       2015,
		}},
		{"9:00 UTCnextweek", &Timespec{
			unit:       IncrementWeeks,
			increments: 1,
			hours:      9,
		}},
	} {
		src := bufio.NewReader(bytes.NewBufferString(testcase.input))
		result := Timespec{}
		err := parseTimespec(src, &result)

		if err != nil {
			t.Fatalf("parseTimespec(%q): %s", testcase.input, err)
		}

		if !reflect.DeepEqual(&result, testcase.expected) {
			t.Fatalf(`parseTimespec(%q):
Expected: %#v
     Got: %#v`)
		}
	}
}
