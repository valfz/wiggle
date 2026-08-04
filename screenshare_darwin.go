//go:build darwin

package main

/*
#include <dlfcn.h>

// CGSIsScreenWatcherPresent (SkyLight/CoreGraphics private API) reports whether
// any process is currently capturing the screen — Zoom, Teams, Meet, QuickTime,
// macOS Screen Sharing, etc. Resolved via dlsym so a missing symbol on a future
// macOS degrades to "not sharing" instead of failing to launch.
static int screenWatcherPresent(void) {
	static int (*fn)(void) = 0;
	if (!fn) {
		void *h = dlopen("/System/Library/PrivateFrameworks/SkyLight.framework/SkyLight", RTLD_LAZY);
		if (h) {
			fn = (int (*)(void))dlsym(h, "SLSIsScreenWatcherPresent");
		}
		if (!fn) {
			fn = (int (*)(void))dlsym(RTLD_DEFAULT, "CGSIsScreenWatcherPresent");
		}
	}
	if (!fn) {
		return 0;
	}
	return fn();
}
*/
import "C"

func screenBeingShared() bool {
	return C.screenWatcherPresent() != 0
}
