package services

import "time"

const (
	asyncCommandTimeout = 5 * time.Minute
	operationTimeout    = time.Hour
	pollingInterval     = 10 * time.Second
)
