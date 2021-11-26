package random

type config struct {
	prefix    []string
	delimiter string
	length    int
}

type Option func(*config)

func WithMaxLength(length int) Option {
	return func(c *config) {
		c.length = length
	}
}

func WithPrefix(prefix ...string) Option {
	return func(c *config) {
		c.prefix = append(c.prefix, prefix...)
	}
}

func WithDelimiter(delimiter string) Option {
	return func(c *config) {
		c.delimiter = delimiter
	}
}

func cfg(opts []Option) (c config) {
	for _, o := range opts {
		o(&c)
	}
	return
}
