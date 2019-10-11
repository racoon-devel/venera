package utils

import (
	"context"
	"math/rand"
	"time"
)

type Range struct {
	MinMs uint32
	MaxMs uint32
}

func Delay(ctx context.Context, delay Range) {
	realDelay := delay.MinMs + rand.Uint32()%(delay.MaxMs-delay.MinMs)
	duration := time.Duration(realDelay) * time.Millisecond
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
		return
	case <-ctx.Done():
		panic("cancelled")
	}
}
