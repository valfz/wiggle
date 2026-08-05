# wiggle

A simple Go tool to keep your computer awake by moving the mouse in a small circle after 90 seconds of inactivity.

## Installation

```sh
go install github.com/valfz/wiggle@latest
```

## Usage

```sh
wiggle start   # run in the background
wiggle stop    # stop the background process
wiggle         # run in the foreground
```

No configuration. The wiggle is skipped while your screen is being shared or recorded (macOS and Windows, best effort).
