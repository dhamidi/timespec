// Package timespec provides functionality for parsing convenient
// definitions of points in time, such as "now next week".
//
// These definitions consist of three parts: a time, a date and an
// increment to add to the specified time.  The date and increment part
// are optional, "now" can be used to indicate the current point in
// time.  Times can be specified in hours (24-hour clock or wall clock),
// optionally followed by minutes. Additionally "noon" is recognized as
// an abbreviation for "12 pm" and "midnight" is an abbreviation for "12
// am".  The following are all valid times: "now", "1 am", "14:15", "1800".
//
// A date can either be a day of the week, such as "Tue" or "Tuesday",
// or a month name followed by a day number and optionally a year.  The
// strings "today" and "tomorrow" are also recognized as dates,
// indicating the obvious.  The following are all valid dates: "Feb 01",
// "today", "Mar 02, 2015", "tomorrow".
//
// Increments are useful for describing points in time relative to a
// reference time such as "now".  An increment is either "+" or the word
// "next", followed by a number and a unit such as "month".  The
// following are all valid increments: "+ 1 year", "next week", "+ 10
// minutes".
//
// The syntax of timespec implemented by this package is the one
// understood by at(1) and reproduced here for convenience:
//
//    %token hr24clock_hr_min
//    %token hr24clock_hour
//    /*
//      An hr24clock_hr_min is a one, two, or four-digit number. A one-digit
//      or two-digit number constitutes an hr24clock_hour. An hr24clock_hour
//      may be any of the single digits [0,9], or may be double digits, ranging
//      from [00,23]. If an hr24clock_hr_min is a four-digit number, the
//      first two digits shall be a valid hr24clock_hour, while the last two
//      represent the number of minutes, from [00,59].
//    */
//
//
//    %token wallclock_hr_min
//    %token wallclock_hour
//    /*
//      A wallclock_hr_min is a one, two-digit, or four-digit number.
//      A one-digit or two-digit number constitutes a wallclock_hour.
//      A wallclock_hour may be any of the single digits [1,9], or may
//      be double digits, ranging from [01,12]. If a wallclock_hr_min
//      is a four-digit number, the first two digits shall be a valid
//      wallclock_hour, while the last two represent the number of
//      minutes, from [00,59].
//    */
//
//
//    %token minute
//    /*
//      A minute is a one or two-digit number whose value can be [0,9]
//      or [00,59].
//    */
//
//
//    %token day_number
//    /*
//      A day_number is a number in the range appropriate for the particular
//      month and year specified by month_name and year_number, respectively.
//      If no year_number is given, the current year is assumed if the given
//      date and time are later this year. If no year_number is given and
//      the date and time have already occurred this year and the month is
//      not the current month, next year is the assumed year.
//    */
//
//
//    %token year_number
//    /*
//      A year_number is a four-digit number representing the year A.D., in
//      which the at_job is to be run.
//    */
//
//
//    %token inc_number
//    /*
//      The inc_number is the number of times the succeeding increment
//      period is to be added to the specified date and time.
//    */
//
//
//    %token timezone_name
//    /*
//      The name of an optional timezone suffix to the time field, in an
//      implementation-defined format.
//    */
//
//
//    %token month_name
//    /*
//      One of the values from the mon or abmon keywords in the LC_TIME
//      locale category.
//    */
//
//
//    %token day_of_week
//    /*
//      One of the values from the day or abday keywords in the LC_TIME
//      locale category.
//    */
//
//
//    %token am_pm
//    /*
//      One of the values from the am_pm keyword in the LC_TIME locale
//      category.
//    */
//
//
//    %start timespec
//    %%
//    timespec    : time
//                | time date
//                | time increment
//                | time date increment
//                | nowspec
//                ;
//
//
//    nowspec     : "now"
//                | "now" increment
//                ;
//
//
//    time        : hr24clock_hr_min
//                | hr24clock_hr_min timezone_name
//                | hr24clock_hour ":" minute
//                | hr24clock_hour ":" minute timezone_name
//                | wallclock_hr_min am_pm
//                | wallclock_hr_min am_pm timezone_name
//                | wallclock_hour ":" minute am_pm
//                | wallclock_hour ":" minute am_pm timezone_name
//                | "noon"
//                | "midnight"
//                ;
//
//
//    date        : month_name day_number
//                | month_name day_number "," year_number
//                | day_of_week
//                | "today"
//                | "tomorrow"
//                ;
//
//
//    increment   : "+" inc_number inc_period
//                | "next" inc_period
//                ;
//
//
//    inc_period  : "minute" | "minutes"
//                | "hour" | "hours"
//                | "day" | "days"
//                | "week" | "weeks"
//                | "month" | "months"
//                | "year" | "years"
//                ;
//
// The only valid timezone_name recognized by this implementation is
// "UTC" (matched case-insensitively).
package timespec

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// A Timespec represents the result of parsing the definition of a point
// in time as understood by at(1).
//
// The point in time described by a Timespec is taken to be in UTC.
type Timespec struct {
	month      time.Month
	day        int
	year       int
	hours      int
	minutes    int
	seconds    int
	isNow      bool
	isTomorrow bool
	increments int
	unit       incrementType
}

// ParseError describes a problem parsing a timespec.
type ParseError struct {
	// Src is the string being parsed
	Src string
	// Pos is offset in bytes at which the error occurred
	Pos int
	// Msg describes the error condition
	Msg string
}

// Error returns the string representation of a ParseError.
func (err *ParseError) Error() string {
	return fmt.Sprintf("at position %d in %q: %s", err.Pos, err.Src, err.Msg)
}

// Parse parses a timespec.
//
// If an error is returned, it is of type *ParseError.
func Parse(timespec string) (*Timespec, error) {
	buf := &buffer{src: timespec, pos: 0}
	spec := &Timespec{}
	err := parseTimespec(buf, spec)

	if err != nil {
		return nil, &ParseError{Src: timespec, Pos: buf.pos, Msg: err.Error()}
	} else {
		return spec, nil
	}
}

// Resolve converts a timespec to a time value, using the provided time
// for resolving "now", "today" and "tomorrow".
//
// The resulting time is in UTC.
func (d *Timespec) Resolve(now time.Time) time.Time {
	if d.isNow {
		d.fromTime(now)
	}

	if d.isTomorrow {
		d.day = d.day + 1
	}

	d.addincrement()

	return time.Date(d.year, d.month, d.day, d.hours, d.minutes, d.seconds, 0, time.UTC)
}

// Time is a convenience function and the same as Resolve(time.Now()).
func (d *Timespec) Time() time.Time {
	return d.Resolve(time.Now())
}

func (d *Timespec) fromTime(t time.Time) {
	d.year, d.month, d.day = t.Date()
	d.hours, d.minutes, d.seconds = t.Clock()
}

func (d *Timespec) isToday() bool {
	return d.year == 0 && d.month == 0 && d.day == 0
}

func (d *Timespec) setToday() {
	d.year = 0
	d.month = 0
	d.day = 0
}

func (d *Timespec) addincrement() {
	switch d.unit {
	case incrementMinutes:
		d.minutes = d.minutes + d.increments
	case incrementHours:
		d.hours = d.hours + d.increments
	case incrementDays:
		d.day = d.day + d.increments
	case incrementWeeks:
		d.day = d.day + 7*d.increments
	case incrementMonths:
		d.month = d.month + time.Month(d.increments)
	case incrementYears:
		d.year = d.year + d.increments
	}
}

// Buffer holds the string to parse.
//
// The only error any methods can return is io.EOF.  Additionally it
// keeps track of the current position in bytes for reporting parsing
// errors.
type buffer struct {
	src string
	pos int
}

func (buf *buffer) ReadByte() (byte, error) {
	if buf.pos >= len(buf.src) {
		return 0, io.EOF
	}

	c := buf.src[buf.pos]

	buf.pos++

	return c, nil
}

func (buf *buffer) UnreadByte() error {
	if buf.pos <= 0 {
		return nil
	}

	buf.pos--

	return nil
}

type incrementType int

const (
	incrementMinutes incrementType = iota
	incrementHours
	incrementDays
	incrementWeeks
	incrementMonths
	incrementYears
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

func peek(in io.ByteScanner) byte {
	c, err := in.ReadByte()

	if err != nil && err != io.EOF {
		panic(err)
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

		buf = append(buf, c)

		if c != expected {
			in.UnreadByte()
			return string(buf), false
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
	c := peek(in)
	if c == 0 {
		return fmt.Errorf("timespec: unexpected EOF")
	}

	if c == 'n' {
		actual, ok := expectBytes(in, []byte("now"))
		if !ok {
			return fmt.Errorf("timespec: expected %q, got %q", "now", actual)
		}

		spec.isNow = true
		return parseincrement(in, spec)
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

	err = parseincrement(in, spec)
	if err != nil {
		spec.increments = 0
	}

	return nil
}

func parseincrement(in io.ByteScanner, spec *Timespec) error {
	skip(in, isspace)
	c, _ := in.ReadByte()

	if c == 0 {
		return nil
	}

	if c == 'n' {
		in.UnreadByte()
		actual, ok := expectBytes(in, []byte("next"))
		if !ok {
			return fmt.Errorf("increment: expected \"next\", got %q", actual)
		}

		spec.increments = 1
	} else if c == '+' {
		buf := []byte{}
		skip(in, isspace)
		any(in, &buf, isdigit)
		count, err := strconv.ParseInt(string(buf), 10, 0)
		if err != nil {
			return fmt.Errorf("increment: %s", err)
		}

		spec.increments = int(count)
	} else {
		return fmt.Errorf("increment: expected '+', got '%c'", c)
	}

	buf := []byte{}
	skip(in, isspace)
	any(in, &buf, nospace)

	period := findPeriod(buf)
	if period == -1 {
		return fmt.Errorf("period: invalid period: %q", buf)
	}

	spec.unit = incrementType(period)

	return nil
}

func findPeriod(buf []byte) int {
	return findInRegexpList(periodNames, buf)
}

func parseDate(in io.ByteScanner, spec *Timespec) error {
	c := peek(in)

	if c == 0 {
		return fmt.Errorf("date: unexpected EOF")
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
		return fmt.Errorf("date: invalid month name: %q", buf)
	}

	spec.month = time.Month(month)

	return parseMonth(in, spec)
}

func parseMonth(in io.ByteScanner, spec *Timespec) error {
	buf := []byte{}
	skip(in, isspace)
	c, ok := expectN(2, in, &buf, isdigit)
	if !ok {
		return fmt.Errorf("month: expected 2 digits, got: %q", buf)
	}

	day, err := strconv.Atoi(string(buf))
	if err != nil {
		return fmt.Errorf("month: invalid day number: %s", buf)
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
		return fmt.Errorf("year: invalid year format: %q", buf)
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
	c := peek(in)

	if isdigit(c) {
		return parseClock(in, spec)
	} else if c == 'n' {
		return parseNoon(in, spec)
	} else if c == 'm' {
		return parseMidnight(in, spec)
	}

	return fmt.Errorf("time: unexpected character %c", c)
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
		return fmt.Errorf("clock: invalid number format: %s", buf)
	}

	if hours > 23 {
		return fmt.Errorf("clock: invalid hours: %d", hours)
	}

	spec.hours = hours

	c = peek(in)

	if c == 0 {
		return fmt.Errorf("clock: unexpected EOF")
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
		return fmt.Errorf("minute: expected ':' or digit, got '%c'", c)
	} else if isdigit(c) {
		in.UnreadByte()
	} else if c != ':' {
		return nil
	}

	buf := []byte{}
	if c, ok := expectN(2, in, &buf, isdigit); !ok {
		return fmt.Errorf("minute: expected digit, got '%c'", c)
	}
	minutes, err := strconv.Atoi(string(buf))
	if err != nil {
		return fmt.Errorf("minute: %s", err)
	}

	if minutes >= 60 {
		return fmt.Errorf("minute: invalid minutes: %d", minutes)
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
		return fmt.Errorf("timezone: invalid timezone: %q", buf)
	}

	return nil
}

func parseAmPm(in io.ByteScanner, spec *Timespec) error {
	c, err := in.ReadByte()
	buf := []byte{c}

	c, err = in.ReadByte()
	if err != nil {
		return fmt.Errorf("am_pm: %s", err)
	}

	if c != 'm' && c != 'M' {
		return fmt.Errorf("am_pm: expected 'm', got %c", c)
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
		return fmt.Errorf("noon: expected %q, got %q", "noon", s)
	}

	spec.hours = 12

	return nil
}

func parseMidnight(in io.ByteScanner, spec *Timespec) error {
	s, ok := expectBytes(in, []byte("midnight"))
	if !ok {
		return fmt.Errorf("midnight: expected %q, got %q", "midnight", s)
	}

	return nil
}
