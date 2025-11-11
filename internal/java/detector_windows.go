//go:build windows

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

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// Windows Registry detection implementation
func (d *Detector) detectFromWindowsRegistry() []Installation {
	var installations []Installation

	// Java Runtime Environment registry paths
	jrePaths := []string{
		`SOFTWARE\JavaSoft\Java Runtime Environment`,
		`SOFTWARE\JavaSoft\JRE`,
	}

	// Java Development Kit registry paths
	jdkPaths := []string{
		`SOFTWARE\JavaSoft\Java Development Kit`,
		`SOFTWARE\JavaSoft\JDK`,
	}

	// Check both HKEY_LOCAL_MACHINE and HKEY_CURRENT_USER
	registryRoots := []registry.Key{
		registry.LOCAL_MACHINE,
		registry.CURRENT_USER,
	}

	for _, root := range registryRoots {
		// Detect JRE installations
		for _, jrePath := range jrePaths {
			installs := d.detectFromRegistryPath(root, jrePath, "JRE")
			installations = append(installations, installs...)
		}

		// Detect JDK installations
		for _, jdkPath := range jdkPaths {
			installs := d.detectFromRegistryPath(root, jdkPath, "JDK")
			installations = append(installations, installs...)
		}
	}

	// Detect modern Java distributions
	modernInstalls := d.detectFromModernRegistry()
	installations = append(installations, modernInstalls...)

	return installations
}

func (d *Detector) detectFromRegistryPath(root registry.Key, regPath, javaType string) []Installation {
	var installations []Installation

	key, err := registry.OpenKey(root, regPath, registry.QUERY_VALUE|registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return installations
	}
	defer key.Close()

	// Get current version
	currentVersion, _, err := key.GetStringValue("CurrentVersion")
	if err != nil {
		currentVersion = ""
	}

	// Enumerate all subkeys (versions)
	subkeys, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return installations
	}

	for _, version := range subkeys {
		versionKey, err := registry.OpenKey(key, version, registry.QUERY_VALUE)
		if err != nil {
			continue
		}

		javaHome, _, err := versionKey.GetStringValue("JavaHome")
		if err != nil {
			versionKey.Close()
			continue
		}
		versionKey.Close()

		// Validate the installation
		if d.isValidJavaHome(javaHome) {
			// Determine if this is the current version
			isCurrent := (version == currentVersion) || (javaHome == d.currentHome)

			// Get more detailed version info
			detailedVersion := d.getJavaVersion(javaHome)
			if detailedVersion == "" || detailedVersion == "不明" {
				detailedVersion = fmt.Sprintf("Java %s", version)
			}

			installations = append(installations, Installation{
				Version: fmt.Sprintf("%s (Registry %s)", detailedVersion, javaType),
				Path:    filepath.Join(javaHome, "bin", d.javaExeName),
				Home:    javaHome,
				Current: isCurrent,
			})
		}
	}

	return installations
}

// Enhanced registry detection for modern Java distributions
func (d *Detector) detectFromModernRegistry() []Installation {
	var installations []Installation

	// Modern Java distributions registry paths
	modernPaths := []string{
		`SOFTWARE\Eclipse Adoptium`,
		`SOFTWARE\Eclipse Foundation`,
		`SOFTWARE\AdoptOpenJDK`,
		`SOFTWARE\Amazon\Corretto`,
		`SOFTWARE\Azul Systems\Zulu`,
		`SOFTWARE\BellSoft\Liberica`,
		`SOFTWARE\Microsoft\OpenJDK`,
	}

	for _, path := range modernPaths {
		key, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.ENUMERATE_SUB_KEYS)
		if err != nil {
			continue
		}

		subkeys, err := key.ReadSubKeyNames(-1)
		if err != nil {
			key.Close()
			continue
		}

		for _, subkey := range subkeys {
			subkeyPath := path + `\` + subkey
			installs := d.detectFromModernDistribution(subkeyPath, subkey)
			installations = append(installations, installs...)
		}

		key.Close()
	}

	return installations
}

func (d *Detector) detectFromModernDistribution(regPath, distName string) []Installation {
	var installations []Installation

	key, err := registry.OpenKey(registry.LOCAL_MACHINE, regPath, registry.QUERY_VALUE|registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return installations
	}
	defer key.Close()

	// Try to get installation path directly
	if installPath, _, err := key.GetStringValue("Path"); err == nil {
		if d.isValidJavaHome(installPath) {
			version := d.getJavaVersion(installPath)
			if version != "" && version != "不明" {
				installations = append(installations, Installation{
					Version: fmt.Sprintf("%s (%s)", version, distName),
					Path:    filepath.Join(installPath, "bin", d.javaExeName),
					Home:    installPath,
					Current: installPath == d.currentHome,
				})
			}
		}
	}

	// Enumerate version subkeys
	subkeys, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return installations
	}

	for _, version := range subkeys {
		versionKey, err := registry.OpenKey(key, version, registry.QUERY_VALUE)
		if err != nil {
			continue
		}

		// Try different common value names
		valueNames := []string{"Path", "InstallationPath", "JavaHome", "Home"}
		var javaHome string

		for _, valueName := range valueNames {
			if path, _, err := versionKey.GetStringValue(valueName); err == nil {
				javaHome = path
				break
			}
		}

		versionKey.Close()

		if javaHome != "" && d.isValidJavaHome(javaHome) {
			detailedVersion := d.getJavaVersion(javaHome)
			if detailedVersion == "" || detailedVersion == "不明" {
				detailedVersion = fmt.Sprintf("Java %s", version)
			}

			installations = append(installations, Installation{
				Version: fmt.Sprintf("%s (%s)", detailedVersion, distName),
				Path:    filepath.Join(javaHome, "bin", d.javaExeName),
				Home:    javaHome,
				Current: javaHome == d.currentHome,
			})
		}
	}

	return installations
}
