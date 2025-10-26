//go:build !change

package tabletest

import (
	"errors"
	"time"
)

var errLeadingInt = errors.New("time: bad [0-9]*") // never printed

// leadingInt consumes the leading [0-9]* from s.
func leadingInt(s string) (x int64, rem string, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > (1<<63-1)/10 {
			// overflow
			return 0, "", errLeadingInt
		}
		x = x*10 + int64(c) - '0'
		if x < 0 {
			// overflow
			return 0, "", errLeadingInt
		}
	}
	return x, s[i:], nil
}

// leadingFraction consumes the leading [0-9]* from s.
// It is used only for fractions, so does not return an error on overflow,
// it just stops accumulating precision.
func leadingFraction(s string) (x int64, scale float64, rem string) {
	i := 0
	scale = 1
	overflow := false
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if overflow {
			continue
		}
		if x > (1<<63-1)/10 {
			// It's possible for overflow to give a positive number, so take care.
			overflow = true
			continue
		}
		y := x*10 + int64(c) - '0'
		if y < 0 {
			overflow = true
			continue
		}
		x = y
		scale *= 10
	}
	return x, scale, s[i:]
}

var unitMap = map[string]int64{
	"ns": int64(time.Nanosecond),
	"us": int64(time.Microsecond),
	"µs": int64(time.Microsecond), // U+00B5 = micro symbol
	"μs": int64(time.Microsecond), // U+03BC = Greek letter mu
	"ms": int64(time.Millisecond),
	"s":  int64(time.Second),
	"m":  int64(time.Minute),
	"h":  int64(time.Hour),
}

// parseNumber parses a number with optional decimal point and returns
// the integer and fractional parts, along with the remaining string.
func parseNumber(s string, orig string) (v, f int64, scale float64, rem string, err error) {
	scale = 1

	// The next character must be [0-9.]
	if s[0] != '.' && (s[0] < '0' || s[0] > '9') {
		return 0, 0, 0, "", errors.New("time: invalid duration " + orig)
	}

	// Consume [0-9]*
	pl := len(s)
	v, s, err = leadingInt(s)
	if err != nil {
		return 0, 0, 0, "", errors.New("time: invalid duration " + orig)
	}
	pre := pl != len(s) // whether we consumed anything before a period

	// Consume (\.[0-9]*)?
	post := false
	if s != "" && s[0] == '.' {
		s = s[1:]
		pl := len(s)
		f, scale, s = leadingFraction(s)
		post = pl != len(s)
	}
	if !pre && !post {
		// no digits (e.g. ".s" or "-.s")
		return 0, 0, 0, "", errors.New("time: invalid duration " + orig)
	}

	return v, f, scale, s, nil
}

// parseUnit extracts the unit string and looks it up in unitMap.
func parseUnit(s string, orig string) (unit int64, rem string, err error) {
	// Consume unit.
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c == '.' || '0' <= c && c <= '9' {
			break
		}
	}
	if i == 0 {
		return 0, "", errors.New("time: missing unit in duration " + orig)
	}
	u := s[:i]
	s = s[i:]
	unit, ok := unitMap[u]
	if !ok {
		return 0, "", errors.New("time: unknown unit " + u + " in duration " + orig)
	}
	return unit, s, nil
}

// computeValue multiplies the parsed number by the unit and checks for overflow.
func computeValue(v, f int64, scale float64, unit int64, orig string) (int64, error) {
	if v > (1<<63-1)/unit {
		// overflow
		return 0, errors.New("time: invalid duration " + orig)
	}
	v *= unit
	if f > 0 {
		// float64 is needed to be nanosecond accurate for fractions of hours.
		// v >= 0 && (f*unit/scale) <= 3.6e+12 (ns/h, h is the largest unit)
		v += int64(float64(f) * (float64(unit) / scale))
		if v < 0 {
			// overflow
			return 0, errors.New("time: invalid duration " + orig)
		}
	}
	return v, nil
}

// ParseDuration parses a duration string.
// A duration string is a possibly signed sequence of
// decimal numbers, each with optional fraction and a unit suffix,
// such as "300ms", "-1.5h" or "2h45m".
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
func ParseDuration(s string) (time.Duration, error) {
	// [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+
	orig := s
	var d int64
	neg := false

	// Consume [-+]?
	if s != "" {
		c := s[0]
		if c == '-' || c == '+' {
			neg = c == '-'
			s = s[1:]
		}
	}
	// Special case: if all that is left is "0", this is zero.
	if s == "0" {
		return 0, nil
	}
	if s == "" {
		return 0, errors.New("time: invalid duration " + orig)
	}
	for s != "" {
		var (
			v, f  int64   // integers before, after decimal point
			scale float64 // value = v + f/scale
			err   error
		)

		// Parse number with optional decimal point
		v, f, scale, s, err = parseNumber(s, orig)
		if err != nil {
			return 0, err
		}

		// Parse unit
		unit, rem, err := parseUnit(s, orig)
		if err != nil {
			return 0, err
		}
		s = rem

		// Compute value with overflow checks
		value, err := computeValue(v, f, scale, unit, orig)
		if err != nil {
			return 0, err
		}

		d += value
		if d < 0 {
			// overflow
			return 0, errors.New("time: invalid duration " + orig)
		}
	}

	if neg {
		d = -d
	}
	return time.Duration(d), nil
}
