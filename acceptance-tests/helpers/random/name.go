package random

import (
	"strings"

	"github.com/Pallinder/go-randomdata"
)

// Name returns a random name that cannot be longer (but may be shorter) than 30
// characters, or a lower limit if specified.
func Name(opts ...Option) string {
	c := cfg(append([]Option{WithMaxLength(100), WithDelimiter("-")}, opts...))

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
