package timespec

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Timespec struct {
	time      *Time
	date      *Date
	increment *Increment
}

type Date struct {
	month      time.Month
	day        int
	year       int
	isToday    bool
	isTomorrow bool
}

type Time struct {
	hours        int
	minutes      int
	timezoneName string
	isNow        bool
}

type IncrementType int

const (
	IncrementMinutes IncrementType = iota
	IncrementHours
	IncrementDays
	IncrementMonths
	IncrementYears
)

type Increment struct {
	count int
	unit  IncrementType
}

var (
	monthNames = []*regexp.Regexp{
		regexp.MustCompile("Jan(uary)?"),
		regexp.MustCompile("Feb(ruary)?"),
		regexp.MustCompile("Mar(ch)?"),
		regexp.MustCompile("Apr(il)?"),
		regexp.MustCompile("May"),
		regexp.MustCompile("June?"),
		regexp.MustCompile("July?"),
		regexp.MustCompile("Aug(ust)?"),
		regexp.MustCompile("Sep(tember)?"),
		regexp.MustCompile("Oct(ober)?"),
		regexp.MustCompile("Nov(ember)?"),
		regexp.MustCompile("Dec(ember)?"),
	}
	dayNames = []*regexp.Regexp{
		regexp.MustCompile("Mon(day)?"),
		regexp.MustCompile("Tue(sday)?"),
		regexp.MustCompile("Wed(nesday)?"),
		regexp.MustCompile("Thu(rsday)?"),
		regexp.MustCompile("Fri(day)?"),
		regexp.MustCompile("Sat(urday)?"),
		regexp.MustCompile("Sun(day)?"),
	}
)

type charclass func(r byte) bool

func isdigit(r byte) bool {
	return r >= '0' && r <= '9'
}

func isspace(r byte) bool {
	return r == ' ' || r == '\n' || r == '\t'
}

func not(class charclass) charclass {
	return func(r byte) bool {
		return !class(r)
	}
}

func skip(in io.ByteScanner, class charclass) byte {
	c, err := in.ReadByte()
	if c == 0 {
		return c
	}
	for class(c) && err == nil {
		c, err = in.ReadByte()
	}

	in.UnreadByte()

	return c
}

func expect(in io.ByteScanner, out *[]byte, class charclass) (byte, bool) {
	c, _ := in.ReadByte()

	if !class(c) {
		in.UnreadByte()
		return c, false
	} else {
		*out = append(*out, c)
		return c, true
	}
}

func expectBytes(in io.ByteScanner, s []byte) (string, bool) {
	buf := []byte{}

	for _, expected := range s {
		c, _ := in.ReadByte()
		if c != expected {
			return string(buf), false
		} else {
			buf = append(buf, c)
		}
	}

	return "", true
}

func any(in io.ByteScanner, out *[]byte, class charclass) {
	c, err := in.ReadByte()
	if err != nil {
		return
	}

	for class(c) && err == nil {
		*out = append(*out, c)
		c, err = in.ReadByte()
	}

	in.UnreadByte()
}

func expectN(n int, in io.ByteScanner, out *[]byte, class charclass) (byte, bool) {
	var c byte

	for i := 0; i < n; i++ {
		c, ok := expect(in, out, class)
		if !ok {
			return c, false
		}
	}

	return c, true
}

func parseDate(in io.ByteScanner, date *Date) error {
	c, _ := in.ReadByte()
	in.UnreadByte()

	if c == 0 {
		return fmt.Errorf("parseDate: unexpected EOF")
	}

	buf := []byte{}
	skip(in, isspace)
	any(in, &buf, not(isspace))

	if string(buf) == "today" {
		date.isToday = true
		return nil
	}

	if string(buf) == "tomorrow" {
		date.isTomorrow = true
		return nil
	}

	day := findDayOfWeek(buf)
	if day != -1 {
		date.day = day
		return nil
	}

	month := findMonth(buf)
	if month == -1 {
		return fmt.Errorf("parseDate: Invalid month name: %q", buf)
	}

	date.month = time.Month(month)

	return parseMonth(in, date)
}

func parseMonth(in io.ByteScanner, date *Date) error {
	buf := []byte{}
	skip(in, isspace)
	c, ok := expectN(2, in, &buf, isdigit)
	if !ok {
		return fmt.Errorf("parseMonth: Expected 2 digits, got: %q", buf)
	}

	day, err := strconv.Atoi(string(buf))
	if err != nil {
		return fmt.Errorf("parseMonth: Invalid day numer: %s", buf)
	}

	date.day = day

	skip(in, isspace)
	c, _ = in.ReadByte()
	if c == ',' {
		return parseYear(in, date)
	} else {
		in.UnreadByte()
	}

	return nil
}

func parseYear(in io.ByteScanner, date *Date) error {
	buf := []byte{}

	skip(in, isspace)
	expectN(4, in, &buf, isdigit)

	year, err := strconv.ParseInt(string(buf), 10, 0)
	if err != nil {
		return fmt.Errorf("parseYear: Invalid year format: %q", buf)
	}

	date.year = int(year)

	return nil
}

func findInRegexpList(list []*regexp.Regexp, buf []byte) int {
	for index, re := range list {
		if re.Match(buf) {
			return index
		}
	}

	return -1
}

func findMonth(buf []byte) int {
	index := findInRegexpList(monthNames, buf)
	if index != -1 {
		return index + 1
	} else {
		return -1
	}
}

func findDayOfWeek(buf []byte) int {
	return findInRegexpList(dayNames, buf)
}

func parseTime(in io.ByteScanner, time *Time) error {
	c, _ := in.ReadByte()
	in.UnreadByte()

	if isdigit(c) {
		return parseClock(in, time)
	} else if c == 'n' {
		return parseNoon(in, time)
	} else if c == 'm' {
		return parseMidnight(in, time)
	}

	return fmt.Errorf("parseTime: Unexpected character %c", c)
}

func parseClock(in io.ByteScanner, time *Time) error {
	c, _ := in.ReadByte()
	buf := []byte{c}

	c, _ = in.ReadByte()

	if c != 0 {
		if !isdigit(c) {
			in.UnreadByte()
			skip(in, isspace)
		} else {
			buf = append(buf, c)
		}
	}

	hours, err := strconv.Atoi(string(buf))
	if err != nil {
		return fmt.Errorf("parseClock: invalid number format: %s", buf)
	}

	if hours > 23 {
		return fmt.Errorf("parseClock: invalid hours: %d", hours)
	}

	time.hours = hours

	c, _ = in.ReadByte()
	in.UnreadByte()

	if c == 0 {
		return fmt.Errorf("parseClock: Unexpected EOF")
	}

	if isdigit(c) || c == ':' {
		if err := parseMinute(in, time); err != nil {
			return err
		}
	}

	c = skip(in, isspace)

	if c != 0 && strings.IndexByte("aApP", c) != -1 {
		if err := parseAmPm(in, time); err != nil {
			return err
		}
	}

	parseTimeZone(in, time)

	return nil
}

func parseMinute(in io.ByteScanner, time *Time) error {
	c, _ := in.ReadByte()

	if c == 0 {
		return nil
	}
	if c != ':' && !isdigit(c) {
		return fmt.Errorf("parseMinute: Expected ':' or digit, got '%c'", c)
	} else if isdigit(c) {
		in.UnreadByte()
	} else if c != ':' {
		return nil
	}

	buf := []byte{}
	if c, ok := expectN(2, in, &buf, isdigit); !ok {
		return fmt.Errorf("parseMinute: Expected digit, got '%c'", c)
	}
	minutes, err := strconv.Atoi(string(buf))
	if err != nil {
		return fmt.Errorf("parseMinute: %s", err)
	}

	if minutes >= 60 {
		return fmt.Errorf("parseMinute: Invalid minutes: %d", minutes)
	}

	time.minutes = minutes

	return nil
}

func parseTimeZone(in io.ByteScanner, time *Time) error {
	buf := []byte{}
	skip(in, isspace)
	any(in, &buf, not(isspace))

	if len(buf) > 0 {
		time.timezoneName = string(buf)
	}

	return nil
}

func parseAmPm(in io.ByteScanner, time *Time) error {
	c, err := in.ReadByte()
	buf := []byte{c}

	c, err = in.ReadByte()
	if err != nil {
		return fmt.Errorf("parseAmPm: %s", err)
	}

	if c != 'm' && c != 'M' {
		return fmt.Errorf("parseAmPm: Expected 'm', got %c", c)
	} else {
		buf = append(buf, c)
	}

	if strings.ToLower(string(buf)) == "pm" {
		time.hours = (time.hours % 12) + 12
	}

	return nil
}

func parseNoon(in io.ByteScanner, time *Time) error {
	s, ok := expectBytes(in, []byte("noon"))
	if !ok {
		return fmt.Errorf("parseNoon: Expected %q, got %q", "noon", s)
	}

	time.hours = 12

	return nil
}

func parseMidnight(in io.ByteScanner, time *Time) error {
	s, ok := expectBytes(in, []byte("midnight"))
	if !ok {
		return fmt.Errorf("parseMidnight: Expected %q, got %q", "midnight", s)
	}

	return nil
}
