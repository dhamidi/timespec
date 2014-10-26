# Description

[![GoDoc](https://godoc.org/github.com/dhamidi/timespec?status.svg)](https://godoc.org/github.com/dhamidi/timespec)

Package timespec parses definitions of points in time, as understood by [POSIX.1-2008 at(1)](http://pubs.opengroup.org/onlinepubs/9699919799/utilities/at.html).  Up-to-date documentation can be found at [godoc](http://godoc.org/github.com/dhamidi/timespec).

Please note that the version of `at(1)` available on Debian extends the
standard (see
[this](http://anonscm.debian.org/cgit/collab-maint/at.git/tree/timespec)
grammar).  The following list shows timespecs that are recognized **only** by Debian's implementation of `at(1)` and **not** understood by this package.

- Dates: `2014-10-26`, `26.10.2014`, `10/26/2014`, `next tuesday`, `26 10 2014`, `Oct 26 2014`
- Times: `teatime`

# Timespecs

A standard timespec consists of three parts: a time, a date and an
increment.  The date and increment parts are optional. The string `now`
is a valid time and date.

A time consists of hours (00 -- 24) and minutes (00 -- 59), optionally
separated by a colon (`:`).  It is also possible to specify in wall
clock format, e.g. `2 pm`.

A date consists either of

- a month name, followed by a day number and optionally a year: `Feb 02` or `Mar 03, 2015`,
- a day of the week: `Tue`, `Monday`,
- or the strings `today` or `tomorrow`.

Increments add a certain amount of time the specified time and date.  A
increment is either a `+`, followed by a number and a unit **or** the word
`next` followed by a unit.  Valid units are `minute`, `hour`, `day`,
`week`, `month`, `year`.  An English plural `s` can be appended for
grammatical reasons, but is not required.
