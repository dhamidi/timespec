package timespec

import (
	"fmt"
	"io"
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
	month time.Month
	day   int
	year  int
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

	time.hours = hours

	isWallClock := (buf[0] == '0' || buf[0] == '1') && hours <= 12

	if isWallClock {
		return parseWallClock(in, time)
	} else {
		return parse24HourClock(in, time)
	}
}

func parseWallClock(in io.ByteScanner, time *Time) error {
	c, _ := in.ReadByte()
	in.UnreadByte()

	if c == 0 {
		return fmt.Errorf("parseWallClock: Unexpected EOF")
	}

	if isdigit(c) || c == ':' {
		if err := parseMinute(in, time); err != nil {
			return err
		}
	} else {
		c = skip(in, isspace)
	}

	if c == 0 {
		return fmt.Errorf("parseWallClock: Expected 'am' or 'pm', got EOF")
	}

	if strings.IndexByte("aApP", c) != -1 {
		if err := parseAmPm(in, time); err != nil {
			return err
		}
	}

	parseTimeZone(in, time)

	return nil
}

func parse24HourClock(in io.ByteScanner, time *Time) error {
	c, _ := in.ReadByte()
	if c == 0 {
		return nil
	}
	in.UnreadByte()

	if isdigit(c) || c == ':' {
		if err := parseMinute(in, time); err != nil {
			return err
		}
	}

	parseTimeZone(in, time)

	return nil
}

func parseMinute(in io.ByteScanner, time *Time) error {
	c, err := in.ReadByte()

	if err != nil {
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

func parseNoon(in io.ByteScanner, time *Time) error     { return nil }
func parseMidnight(in io.ByteScanner, time *Time) error { return nil }
