PACKAGE DOCUMENTATION

package timespec
    import "github.com/dhamidi/timespec"

    Package timespec provides functionality for parsing convenient
    definitions of points in time, such as "now next week".

    These definitions consist of three parts: a time, a date and an
    increment to add to the specified time. The date and increment part are
    optional, "now" can be used to indicate the current point in time. Times
    can be specified in hours (24-hour clock or wall clock), optionally
    followed by minutes. Additionally "noon" is recognized as an
    abbreviation for "12 pm" and "midnight" is an abbreviation for "12 am".
    The following are all valid times: "now", "1 am", "14:15", "1800".

    A date can either be a day of the week, such as "Tue" or "Tuesday", or a
    month name followed by a day number and optionally a year. The strings
    "today" and "tomorrow" are also recognized as dates, indicating the
    obvious. The following are all valid dates: "Feb 01", "today", "Mar 02,
    2015", "tomorrow".

    Increments are useful for describing points in time relative to a
    reference time such as "now". An increment is either "+" or the word
    "next", followed by a number and a unit such as "month". The following
    are all valid increments: "+ 1 year", "next week", "+ 10 minutes".

    The syntax of timespec implemented by this package is the one understood
    by at(1) and reproduced here for convenience:

	%token hr24clock_hr_min
	%token hr24clock_hour
	/*
	  An hr24clock_hr_min is a one, two, or four-digit number. A one-digit
	  or two-digit number constitutes an hr24clock_hour. An hr24clock_hour
	  may be any of the single digits [0,9], or may be double digits, ranging
	  from [00,23]. If an hr24clock_hr_min is a four-digit number, the
	  first two digits shall be a valid hr24clock_hour, while the last two
	  represent the number of minutes, from [00,59].
	*/

	%token wallclock_hr_min
	%token wallclock_hour
	/*
	  A wallclock_hr_min is a one, two-digit, or four-digit number.
	  A one-digit or two-digit number constitutes a wallclock_hour.
	  A wallclock_hour may be any of the single digits [1,9], or may
	  be double digits, ranging from [01,12]. If a wallclock_hr_min
	  is a four-digit number, the first two digits shall be a valid
	  wallclock_hour, while the last two represent the number of
	  minutes, from [00,59].
	*/

	%token minute
	/*
	  A minute is a one or two-digit number whose value can be [0,9]
	  or [00,59].
	*/

	%token day_number
	/*
	  A day_number is a number in the range appropriate for the particular
	  month and year specified by month_name and year_number, respectively.
	  If no year_number is given, the current year is assumed if the given
	  date and time are later this year. If no year_number is given and
	  the date and time have already occurred this year and the month is
	  not the current month, next year is the assumed year.
	*/

	%token year_number
	/*
	  A year_number is a four-digit number representing the year A.D., in
	  which the at_job is to be run.
	*/

	%token inc_number
	/*
	  The inc_number is the number of times the succeeding increment
	  period is to be added to the specified date and time.
	*/

	%token timezone_name
	/*
	  The name of an optional timezone suffix to the time field, in an
	  implementation-defined format.
	*/

	%token month_name
	/*
	  One of the values from the mon or abmon keywords in the LC_TIME
	  locale category.
	*/

	%token day_of_week
	/*
	  One of the values from the day or abday keywords in the LC_TIME
	  locale category.
	*/

	%token am_pm
	/*
	  One of the values from the am_pm keyword in the LC_TIME locale
	  category.
	*/

	%start timespec
	%%
	timespec    : time
	            | time date
	            | time increment
	            | time date increment
	            | nowspec
	            ;

	nowspec     : "now"
	            | "now" increment
	            ;

	time        : hr24clock_hr_min
	            | hr24clock_hr_min timezone_name
	            | hr24clock_hour ":" minute
	            | hr24clock_hour ":" minute timezone_name
	            | wallclock_hr_min am_pm
	            | wallclock_hr_min am_pm timezone_name
	            | wallclock_hour ":" minute am_pm
	            | wallclock_hour ":" minute am_pm timezone_name
	            | "noon"
	            | "midnight"
	            ;

	date        : month_name day_number
	            | month_name day_number "," year_number
	            | day_of_week
	            | "today"
	            | "tomorrow"
	            ;

	increment   : "+" inc_number inc_period
	            | "next" inc_period
	            ;

	inc_period  : "minute" | "minutes"
	            | "hour" | "hours"
	            | "day" | "days"
	            | "week" | "weeks"
	            | "month" | "months"
	            | "year" | "years"
	            ;

    The only valid timezone_name recognized by this implementation is "UTC"
    (matched case-insensitively).

TYPES

type ParseError struct {
    // Src is the string being parsed
    Src string
    // Pos is offset in bytes at which the error occurred
    Pos int
    // Msg describes the error condition
    Msg string
}
    ParseError describes a problem parsing a timespec.

func (err *ParseError) Error() string
    Error returns the string representation of a ParseError.

type Timespec struct {
    // contains filtered or unexported fields
}
    A Timespec represents the result of parsing the definition of a point in
    time as understood by at(1).

    The point in time described by a Timespec is taken to be in UTC.

func Parse(timespec string) (*Timespec, error)
    Parse parses a timespec.

    If an error is returned, it is of type *ParseError.

    Example:
	now := time.Date(2010, 1, 1, 12, 0, 0, 0, time.UTC)
	spec, err := Parse("now next week")
	
	if err != nil {
	    fmt.Println(err)
	    return
	}
	
	fmt.Println(spec.Resolve(now))
	// Output: 2010-01-08 12:00:00 +0000 UTC
func (d *Timespec) Resolve(now time.Time) time.Time
    Resolve converts a timespec to a time value, using the provided time for
    resolving "now", "today" and "tomorrow".

    The resulting time is in UTC.

func (d *Timespec) Time() time.Time
    Time is a convenience function and the same as Resolve(time.Now()).


