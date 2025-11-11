//go:build unix

/*
 * JSwitcher
 *
 * Copyright 2025 s12kuma01
 *
 * This software is licensed under the Open Software License version
 * 3.0. The full text of this license can be found in https://opensource.org/licenses/OSL-3.0
 * or in the LICENSES directory which is distributed along with the software.
 */

package java

// Non-Windows stub for registry detection
func (d *Detector) detectFromWindowsRegistry() []Installation {
	return []Installation{}
}
