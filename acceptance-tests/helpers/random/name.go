package random

import (
	"strings"

	"github.com/Pallinder/go-randomdata"
)

// We have seen occasions where the same name has been returned more than once, so
// we should keep a record of previous values to avoid duplicates
var previous map[string]struct{}

// Name returns a random name that cannot be longer (but may be shorter) than 30
// characters, or a lower limit if specified.
func Name(opts ...Option) string {
	c := cfg(append([]Option{WithMaxLength(100), WithDelimiter("-")}, opts...))

	generate := func() string {
		parts := c.prefix

		// Do we have enough available length to add an adjective?
		if c.length > len(strings.Join(parts, c.delimiter))+20 {
			parts = append(parts, randomdata.Adjective())
		}
		parts = append(parts, randomdata.Noun())

		joined := strings.Join(parts, c.delimiter)
		if len(joined) > c.length {
			return joined[:c.length]
		}

		return joined
	}

	if previous == nil {
		previous = make(map[string]struct{})
	}

	for {
		value := generate()
		if _, ok := previous[value]; !ok {
			previous[value] = struct{}{}
			return value
		}
	}
}
