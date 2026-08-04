//go:build windows

package main

import "golang.org/x/sys/windows/registry"

// Windows tracks apps using the Graphics Capture API in the
// CapabilityAccessManager consent store. An entry with LastUsedTimeStart set
// and LastUsedTimeStop == 0 is capturing the screen right now (new Teams and
// other WGC-based apps register here; DXGI duplication capture does not).
func screenBeingShared() bool {
	const base = `Software\Microsoft\Windows\CurrentVersion\CapabilityAccessManager\ConsentStore\`
	for _, store := range []string{"graphicsCaptureProgrammatic", "graphicsCaptureWithoutBorder"} {
		if storeInUse(base + store) {
			return true
		}
	}
	return false
}

func storeInUse(path string) bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, path, registry.READ)
	if err != nil {
		return false
	}
	defer k.Close()

	subs, err := k.ReadSubKeyNames(-1)
	if err != nil {
		return false
	}
	for _, sub := range subs {
		if sub == "NonPackaged" {
			// Win32 apps live one level deeper, keyed by exe path.
			nk, err := registry.OpenKey(k, sub, registry.READ)
			if err != nil {
				continue
			}
			nsubs, _ := nk.ReadSubKeyNames(-1)
			active := false
			for _, ns := range nsubs {
				if captureActive(nk, ns) {
					active = true
					break
				}
			}
			nk.Close()
			if active {
				return true
			}
			continue
		}
		if captureActive(k, sub) {
			return true
		}
	}
	return false
}

func captureActive(parent registry.Key, name string) bool {
	k, err := registry.OpenKey(parent, name, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	start, _, err := k.GetIntegerValue("LastUsedTimeStart")
	if err != nil || start == 0 {
		return false
	}
	stop, _, err := k.GetIntegerValue("LastUsedTimeStop")
	return err == nil && stop == 0
}
