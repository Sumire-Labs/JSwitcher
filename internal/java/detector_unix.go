//go:build !windows

package java

// Non-Windows stub for registry detection
func (d *Detector) detectFromWindowsRegistry() []Installation {
	return []Installation{}
}