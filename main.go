package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

const (
	IdleTimeout    = 90 * time.Second
	WiggleDuration = 1 * time.Second
	PollInterval   = 250 * time.Millisecond // min 25ms
	Amplitude      = 3                      // Wiggle circle radius in px: Min 1
	WiggleStep     = 10 * time.Millisecond  // Speed of wiggle movement
)

func main() {
	cmd := ""
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	var err error
	switch cmd {
	case "", "run":
		runWiggler()
		return
	case "start":
		err = startDaemon()
	case "stop":
		err = stopDaemon()
	default:
		fmt.Fprintln(os.Stderr, "usage: wiggle [start|stop]")
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "wiggle:", err)
		os.Exit(1)
	}
}

func runWiggler() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Printf("Mouse Wiggler started. Idle timeout: %v, Wiggle: %v, Poll: %v, Amplitude: %dpx\n",
		IdleTimeout, WiggleDuration, PollInterval, Amplitude)
	fmt.Println("Press Ctrl+C to exit.")

	eventCh := make(chan struct{}, 1)
	stopCh := make(chan struct{})
	startEventMonitor(eventCh, stopCh)
	monitor(ctx, eventCh)
	close(stopCh)
}

func startEventMonitor(eventCh chan<- struct{}, stopCh <-chan struct{}) {
	go func() {
		fmt.Println("hook start...")
		evChan := hook.Start()
		defer hook.End()
		for {
			select {
			case _, ok := <-evChan:
				if !ok {
					return
				}
				// Any keyboard or mouse event resets idle timer
				select {
				case eventCh <- struct{}{}:
				default:
				}
			case <-stopCh:
				return
			}
		}
	}()
}

func monitor(ctx context.Context, eventCh <-chan struct{}) {
	xPrev, yPrev := robotgo.Location()
	lastMove := time.Now()
	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			x, y := robotgo.Location()
			if x != xPrev || y != yPrev {
				lastMove = time.Now()
				xPrev, yPrev = x, y
				continue
			}
			if time.Since(lastMove) >= IdleTimeout {
				if screenBeingShared() {
					fmt.Println("screen sharing detected, skipping wiggle")
					lastMove = time.Now()
					continue
				}
				wiggleOnce(ctx)
				xPrev, yPrev = robotgo.Location()
				lastMove = time.Now()
			}
		case <-eventCh:
			lastMove = time.Now()
		}
	}
}

func wiggleOnce(ctx context.Context) {
	ox, oy := robotgo.Location()
	deadline := time.Now().Add(WiggleDuration)

	const numberOfPointsOnCircle = 16
	positions := make([]struct{ dx, dy int }, numberOfPointsOnCircle)
	for i := 0; i < numberOfPointsOnCircle; i++ {
		angle := 2 * 3.14159265 * float64(i) / float64(numberOfPointsOnCircle)
		positions[i].dx = int(float64(Amplitude) * math.Cos(angle))
		positions[i].dy = int(float64(Amplitude) * math.Sin(angle))
	}

	for idx := 0; time.Now().Before(deadline); {
		select {
		case <-ctx.Done():
			robotgo.Move(ox, oy)
			return
		case <-time.After(WiggleStep):
			p := positions[idx]
			robotgo.Move(ox+p.dx, oy+p.dy)
			idx = (idx + 1) % len(positions)
		}
	}
	robotgo.Move(ox, oy)
}
