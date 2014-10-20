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
	month      time.Month
	day        int
	year       int
	hours      int
	minutes    int
	timezone   string
	isNow      bool
	isTomorrow bool
	increments int
	unit       incrementType
}

func (d *Timespec) isToday() bool {
	return d.year == 0 && d.month == 0 && d.day == 0
}

func (d *Timespec) setToday() {
	d.year = 0
	d.month = 0
	d.day = 0
}

type incrementType int

const (
	IncrementMinutes incrementType = iota
	IncrementHours
	IncrementDays
	IncrementWeeks
	IncrementMonths
	IncrementYears
)

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
	periodNames = []*regexp.Regexp{
		regexp.MustCompile("minutes?"),
		regexp.MustCompile("hours?"),
		regexp.MustCompile("days?"),
		regexp.MustCompile("weeks?"),
		regexp.MustCompile("months?"),
		regexp.MustCompile("years?"),
	}
)

type charclass func(r byte) bool

func isdigit(r byte) bool {
	return r >= '0' && r <= '9'
}

func isspace(r byte) bool {
	return r == ' ' || r == '\n' || r == '\t'
}

func nospace(r byte) bool {
	return !isspace(r)
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

func parseTimespec(in io.ByteScanner, spec *Timespec) error {
	c, _ := in.ReadByte()
	in.UnreadByte()
	if c == 0 {
		return fmt.Errorf("parseTimespec: unexpected EOF")
	}

	if c == 'n' {
		actual, ok := expectBytes(in, []byte("now"))
		if !ok {
			return fmt.Errorf("parseTimespec: expected %q, got %q", "now", actual)
		}

		spec.isNow = true
		return parseIncrement(in, spec)
	}

	err := parseTime(in, spec)
	if err != nil {
		return err
	}

	err = parseDate(in, spec)
	if err != nil {
		spec.year = 0
		spec.month = 0
		spec.day = 0
	}

	err = parseIncrement(in, spec)
	if err != nil {
		spec.increments = 0
	}

	return nil
}

func parseIncrement(in io.ByteScanner, spec *Timespec) error {
	skip(in, isspace)
	c, _ := in.ReadByte()

	if c == 0 {
		return nil
	}

	if c == 'n' {
		in.UnreadByte()
		actual, ok := expectBytes(in, []byte("next"))
		if !ok {
			return fmt.Errorf("parseIncrement: Expected \"next\", got %q", actual)
		}

		spec.increments = 1
	} else if c == '+' {
		buf := []byte{}
		skip(in, isspace)
		any(in, &buf, isdigit)
		count, err := strconv.ParseInt(string(buf), 10, 0)
		if err != nil {
			return fmt.Errorf("parseIncrement: %s", err)
		}

		spec.increments = int(count)
	} else {
		return fmt.Errorf("parseIncrement: Expected '+', got '%c'", c)
	}

	buf := []byte{}
	skip(in, isspace)
	any(in, &buf, nospace)

	period := findPeriod(buf)
	if period == -1 {
		return fmt.Errorf("parsePeriod: Invalid period: %q", buf)
	}

	spec.unit = incrementType(period)

	return nil
}

func findPeriod(buf []byte) int {
	return findInRegexpList(periodNames, buf)
}

func parseDate(in io.ByteScanner, spec *Timespec) error {
	c, _ := in.ReadByte()
	in.UnreadByte()

	if c == 0 {
		return fmt.Errorf("parseDate: unexpected EOF")
	}

	buf := []byte{}
	c = skip(in, isspace)
	if c == '+' || c == 'n' {
		return nil
	}

	any(in, &buf, nospace)

	if string(buf) == "today" {
		spec.setToday()
		return nil
	}

	if string(buf) == "tomorrow" {
		spec.isTomorrow = true
		return nil
	}

	day := findDayOfWeek(buf)
	if day != -1 {
		spec.day = day
		return nil
	}

	month := findMonth(buf)
	if month == -1 {
		return fmt.Errorf("parseDate: Invalid month name: %q", buf)
	}

	spec.month = time.Month(month)

	return parseMonth(in, spec)
}

func parseMonth(in io.ByteScanner, spec *Timespec) error {
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

	spec.day = day

	skip(in, isspace)
	c, _ = in.ReadByte()
	if c == ',' {
		return parseYear(in, spec)
	} else {
		in.UnreadByte()
	}

	return nil
}

func parseYear(in io.ByteScanner, spec *Timespec) error {
	buf := []byte{}

	skip(in, isspace)
	expectN(4, in, &buf, isdigit)

	year, err := strconv.ParseInt(string(buf), 10, 0)
	if err != nil {
		return fmt.Errorf("parseYear: Invalid year format: %q", buf)
	}

	spec.year = int(year)

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

func parseTime(in io.ByteScanner, spec *Timespec) error {
	c, _ := in.ReadByte()
	in.UnreadByte()

	if isdigit(c) {
		return parseClock(in, spec)
	} else if c == 'n' {
		return parseNoon(in, spec)
	} else if c == 'm' {
		return parseMidnight(in, spec)
	}

	return fmt.Errorf("parseTime: Unexpected character %c", c)
}

func parseClock(in io.ByteScanner, spec *Timespec) error {
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

	spec.hours = hours

	c, _ = in.ReadByte()
	in.UnreadByte()

	if c == 0 {
		return fmt.Errorf("parseClock: Unexpected EOF")
	}

	if isdigit(c) || c == ':' {
		if err := parseMinute(in, spec); err != nil {
			return err
		}
	}

	c = skip(in, isspace)

	if c != 0 && strings.IndexByte("aApP", c) != -1 {
		if err := parseAmPm(in, spec); err != nil {
			return err
		}
	}

	parseTimeZone(in, spec)

	return nil
}

func parseMinute(in io.ByteScanner, spec *Timespec) error {
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

	spec.minutes = minutes

	return nil
}

func parseTimeZone(in io.ByteScanner, spec *Timespec) error {
	c := skip(in, isspace)
	// only UTC (case insensitive) is a valid timezone
	if c != 'u' && c != 'U' {
		return nil
	}

	buf := []byte{}

	expectN(3, in, &buf, nospace)

	timezone := strings.ToUpper(string(buf))

	if timezone != "UTC" {
		return fmt.Errorf("parseTimeZone: Invalid timezone: %q", buf)
	}

	return nil
}

func parseAmPm(in io.ByteScanner, spec *Timespec) error {
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
		spec.hours = (spec.hours % 12) + 12
	}

	return nil
}

func parseNoon(in io.ByteScanner, spec *Timespec) error {
	s, ok := expectBytes(in, []byte("noon"))
	if !ok {
		return fmt.Errorf("parseNoon: Expected %q, got %q", "noon", s)
	}

	spec.hours = 12

	return nil
}

func parseMidnight(in io.ByteScanner, spec *Timespec) error {
	s, ok := expectBytes(in, []byte("midnight"))
	if !ok {
		return fmt.Errorf("parseMidnight: Expected %q, got %q", "midnight", s)
	}

	return nil
}
