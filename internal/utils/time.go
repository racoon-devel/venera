package utils

import (
	"context"
	"math/rand"
	"time"
)

type Range struct {
	Min time.Duration
	Max time.Duration
}

func Delay(ctx context.Context, delay Range) {
	duration := delay.Min + time.Duration(rand.Int63())%(delay.Max-delay.Min)
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
		return
	case <-ctx.Done():
		panic("cancelled")
	}
}
