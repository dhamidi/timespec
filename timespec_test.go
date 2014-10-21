package timespec

import (
	"bufio"
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"
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
			t.Logf("parseTime(%q): %s", testcase.input, err)
			t.Fail()
		}

		if !reflect.DeepEqual(&result, testcase.expected) {
			t.Logf("parseTime(%q):\n  Expected: %#v\n       Got: %#v\n",
				testcase.input, testcase.expected, &result)
			t.Fail()
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
			t.Logf("parseDate(%q): %s", testcase.input, err)
			t.Fail()
		}

		if !reflect.DeepEqual(&result, testcase.expected) {
			t.Logf("parseDate(%q):\n  Expected: %#v\n       Got: %#v\n",
				testcase.input, testcase.expected, &result)
			t.Fail()
		}

	}
}

func TestParseincrement(t *testing.T) {
	for _, testcase := range []*testTimespec{
		{"+1 day", &Timespec{increments: 1, unit: incrementDays}},
		{"+ 1 day", &Timespec{increments: 1, unit: incrementDays}},
		{"next week", &Timespec{increments: 1, unit: incrementWeeks}},
		{"nextday", &Timespec{increments: 1, unit: incrementDays}},
		{"+ 20 months", &Timespec{increments: 20, unit: incrementMonths}},
	} {
		src := bufio.NewReader(bytes.NewBufferString(testcase.input))
		result := Timespec{}
		err := parseincrement(src, &result)

		if err != nil {
			t.Logf("parseincrement(%q): %s", testcase.input, err)
			t.Fail()
		}

		if !reflect.DeepEqual(&result, testcase.expected) {
			t.Logf("parseincrement(%q):\n  Expected: %#v\n       Got: %#v\n",
				testcase.input, testcase.expected, &result)
			t.Fail()
		}

	}
}

func TestParseTimespec(t *testing.T) {
	for _, testcase := range []*testTimespec{
		{"now + 1 day", &Timespec{
			increments: 1,
			unit:       incrementDays,
			isNow:      true,
		}},
		{"now", &Timespec{isNow: true}},
		{"12:11", &Timespec{
			hours:   12,
			minutes: 11,
		}},
		{"10 am next week", &Timespec{
			increments: 1,
			unit:       incrementWeeks,
			hours:      10,
		}},
		{"14:00 Feb 12, 2015 + 3 week", &Timespec{
			increments: 3,
			unit:       incrementWeeks,
			hours:      14,
			month:      2,
			day:        12,
			year:       2015,
		}},
		{"9:00 UTCnextweek", &Timespec{
			unit:       incrementWeeks,
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

func TestTimespec_Resolve(t *testing.T) {
	now := time.Date(2010, 1, 1, 15, 10, 0, 0, time.UTC)

	testcases := []struct {
		at   Timespec
		then time.Time
	}{
		{
			then: time.Date(2010, 1, 2, 15, 10, 0, 0, time.UTC),
			at:   Timespec{isNow: true, increments: 1, unit: incrementDays},
		},
		{
			then: time.Date(2010, 2, 5, 15, 10, 0, 0, time.UTC),
			at:   Timespec{isNow: true, increments: 5, unit: incrementWeeks},
		},
		{
			then: time.Date(2010, 2, 1, 15, 10, 0, 0, time.UTC),
			at:   Timespec{isNow: true, increments: 1, unit: incrementMonths},
		},
		{
			then: time.Date(2014, 1, 1, 15, 10, 0, 0, time.UTC),
			at:   Timespec{isNow: true, increments: 4, unit: incrementYears},
		},
		{
			then: time.Date(2010, 2, 2, 15, 10, 0, 0, time.UTC),
			at:   Timespec{year: 2010, month: 2, day: 1, hours: 15, minutes: 10, isTomorrow: true},
		},
	}

	for i, testcase := range testcases {
		if resolved := testcase.at.Resolve(now); !resolved.Equal(testcase.then) {
			t.Logf("test[%d]: expected %s to equal %s", i, resolved, testcase.then)
			t.Fail()
		}
	}
}

func ExampleParse() {
	now := time.Date(2010, 1, 1, 12, 0, 0, 0, time.UTC)
	spec, err := Parse("now next week")

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(spec.Resolve(now))
	// Output: 2010-01-08 12:00:00 +0000 UTC
}

func TestTimespec_Resolve_keepsSeconds(t *testing.T) {
	now := time.Date(2010, 1, 1, 15, 10, 23, 0, time.UTC)
	at := &Timespec{isNow: true, increments: 1, unit: incrementDays}
	then := time.Date(2010, 1, 2, 15, 10, 23, 0, time.UTC)

	if atTime := at.Resolve(now); !atTime.Equal(then) {
		t.Fatalf("Expected %s to equal %s", atTime, then)
	}
}
