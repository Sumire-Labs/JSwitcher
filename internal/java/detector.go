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
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

type Installation struct {
	Version string
	Path    string
	Home    string
	Current bool
}

type Detector struct {
	searchPaths []string
	javaExeName string
	currentHome string
}

func NewDetector() *Detector {
	detector := &Detector{
		currentHome: os.Getenv("JAVA_HOME"),
	}

	// Platform-specific configuration
	if strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") {
		detector.searchPaths = []string{
			"C:\\Program Files\\Java",
			"C:\\Program Files (x86)\\Java",
			"C:\\Program Files\\OpenJDK",
			"C:\\Program Files (x86)\\OpenJDK",
			"C:\\Program Files\\Eclipse Adoptium",
			"C:\\Program Files (x86)\\Eclipse Adoptium",
			"C:\\Program Files\\Amazon Corretto",
			"C:\\Program Files (x86)\\Amazon Corretto",
			"C:\\Program Files\\Zulu",
			"C:\\Program Files (x86)\\Zulu",
		}
		detector.javaExeName = "java.exe"
	} else {
		detector.searchPaths = []string{
			"/usr/lib/jvm",
			"/usr/java",
			"/opt/java",
			"/Library/Java/JavaVirtualMachines", // macOS
			"/usr/local/java",
		}
		detector.javaExeName = "java"
	}

	return detector
}

func (d *Detector) DetectInstallations() ([]Installation, error) {
	var installations []Installation

	// 1. Windows Registry detection (highest priority)
	if runtime.GOOS == "windows" {
		registryInstalls := d.detectFromRegistry()
		installations = append(installations, registryInstalls...)
	}

	// 2. SDKMAN detection (Unix systems)
	if runtime.GOOS != "windows" {
		sdkmanInstalls := d.detectFromSDKMAN()
		installations = append(installations, sdkmanInstalls...)
	}

	// 3. Package manager detection
	pkgMgrInstalls := d.detectFromPackageManagers()
	installations = append(installations, pkgMgrInstalls...)

	// 4. PATH environment detection
	pathInstalls := d.detectFromPATH()
	installations = append(installations, pathInstalls...)

	// 5. Traditional directory search (fallback)
	dirInstalls := d.detectFromDirectories()
	installations = append(installations, dirInstalls...)

	// Remove duplicates and sort
	installations = d.removeDuplicates(installations)

	// 最終的なCurrentフラグの確定（JAVA_HOMEと完全一致するもの1つだけ）
	installations = d.ensureSingleCurrent(installations)

	d.sortInstallations(installations)

	return installations, nil
}

func (d *Detector) isJavaDirectory(name string) bool {
	name = strings.ToLower(name)
	return strings.Contains(name, "jdk") ||
		strings.Contains(name, "jre") ||
		strings.Contains(name, "java-") ||
		strings.Contains(name, "corretto") ||
		strings.Contains(name, "zulu") ||
		strings.Contains(name, "adoptium")
}

func (d *Detector) containsInstallation(installations []Installation, home string) bool {
	for _, existing := range installations {
		if existing.Home == home {
			return true
		}
	}
	return false
}

func (d *Detector) sortInstallations(installations []Installation) {
	sort.Slice(installations, func(i, j int) bool {
		if installations[i].Current && !installations[j].Current {
			return true
		}
		if !installations[i].Current && installations[j].Current {
			return false
		}
		return installations[i].Version > installations[j].Version
	})
}

func (d *Detector) extractJavaHomeFromExe(javaPath string) string {
	dir := filepath.Dir(javaPath)
	if strings.HasSuffix(dir, "bin") {
		return filepath.Dir(dir)
	}
	return ""
}

func (d *Detector) getJavaVersion(javaHome string) string {
	javaExe := filepath.Join(javaHome, "bin", d.javaExeName)
	if _, err := os.Stat(javaExe); os.IsNotExist(err) {
		// Try alternative executable name
		if d.javaExeName == "java.exe" {
			javaExe = filepath.Join(javaHome, "bin", "java")
		} else {
			javaExe = filepath.Join(javaHome, "bin", "java.exe")
		}
		if _, err := os.Stat(javaExe); os.IsNotExist(err) {
			return "不明"
		}
	}

	cmd := exec.Command(javaExe, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try to extract version from directory name as fallback
		dirName := filepath.Base(javaHome)
		if versionFromDir := extractVersionFromDirName(dirName); versionFromDir != "" {
			return versionFromDir
		}
		return "不明"
	}

	// Parse version from output
	versionRegex := regexp.MustCompile(`version "([^"]+)"`)
	matches := versionRegex.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		version := matches[1]
		return formatJavaVersion(version)
	}

	// Alternative regex for different output formats
	altRegex := regexp.MustCompile(`(\d+(?:\.\d+)*(?:[._]\d+)*(?:-[a-zA-Z0-9]+)*)`)
	matches = altRegex.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		return formatJavaVersion(matches[1])
	}

	return "Unknown"
}

func extractVersionFromDirName(dirName string) string {
	// Common patterns in directory names
	patterns := []string{
		`jdk-?(\d+(?:\.\d+)*(?:[._]\d+)*)`,
		`jre-?(\d+(?:\.\d+)*(?:[._]\d+)*)`,
		`java-(\d+(?:\.\d+)*(?:[._]\d+)*)`,
		`(\d+(?:\.\d+)*(?:[._]\d+)*)-jdk`,
		`(\d+(?:\.\d+)*(?:[._]\d+)*)-jre`,
	}

	for _, pattern := range patterns {
		regex := regexp.MustCompile(`(?i)` + pattern)
		matches := regex.FindStringSubmatch(dirName)
		if len(matches) > 1 {
			return formatJavaVersion(matches[1])
		}
	}

	return ""
}

func formatJavaVersion(version string) string {
	// Clean version string
	version = strings.ReplaceAll(version, "_", ".")

	// Extract major version for modern Java (9+)
	if strings.HasPrefix(version, "1.") {
		// Java 8 and below (1.8.0_xxx format)
		parts := strings.Split(version, ".")
		if len(parts) >= 2 {
			return "Java " + parts[1]
		}
	} else {
		// Java 9+ (11.0.xx format)
		parts := strings.Split(version, ".")
		if len(parts) >= 1 {
			majorVersion := strings.Split(parts[0], "-")[0]
			majorVersion = strings.Split(majorVersion, "_")[0]
			return "Java " + majorVersion
		}
	}

	return "Java " + version
}

// Windows Registry detection
func (d *Detector) detectFromRegistry() []Installation {
	if runtime.GOOS != "windows" {
		return []Installation{}
	}

	// Note: この関数は条件付きコンパイルで Windows でのみ利用可能
	return d.detectFromWindowsRegistry()
}

// SDKMAN detection for Unix systems
func (d *Detector) detectFromSDKMAN() []Installation {
	var installations []Installation

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return installations
	}

	sdkmanDir := filepath.Join(homeDir, ".sdkman", "candidates", "java")
	if _, err := os.Stat(sdkmanDir); os.IsNotExist(err) {
		return installations
	}

	entries, err := os.ReadDir(sdkmanDir)
	if err != nil {
		return installations
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "current" {
			javaHome := filepath.Join(sdkmanDir, entry.Name())
			if d.isValidJavaHome(javaHome) {
				version := d.getJavaVersion(javaHome)
				if version != "" && version != "不明" {
					installations = append(installations, Installation{
						Version: version + " (SDKMAN)",
						Path:    filepath.Join(javaHome, "bin", d.javaExeName),
						Home:    javaHome,
						Current: javaHome == d.currentHome,
					})
				}
			}
		}
	}

	return installations
}

// Package manager detection
func (d *Detector) detectFromPackageManagers() []Installation {
	var installations []Installation

	switch runtime.GOOS {
	case "darwin":
		// Homebrew detection
		installations = append(installations, d.detectFromHomebrew()...)
	case "linux":
		// APT, YUM detection can be added here
	case "windows":
		// Chocolatey, Scoop detection
		installations = append(installations, d.detectFromChocolatey()...)
		installations = append(installations, d.detectFromScoop()...)
	}

	return installations
}

// Enhanced PATH detection
func (d *Detector) detectFromPATH() []Installation {
	var installations []Installation

	// Find all java executables in PATH
	javaExecutables := d.findAllJavaInPATH()

	for _, javaPath := range javaExecutables {
		if javaHome := d.extractJavaHomeFromExe(javaPath); javaHome != "" {
			version := d.getJavaVersion(javaHome)
			if version != "" && version != "不明" {
				installations = append(installations, Installation{
					Version: version + " (PATH)",
					Path:    javaPath,
					Home:    javaHome,
					Current: javaHome == d.currentHome,
				})
			}
		}
	}

	return installations
}

// Traditional directory search
func (d *Detector) detectFromDirectories() []Installation {
	var installations []Installation

	// Original directory search logic
	for _, searchPath := range d.searchPaths {
		if _, err := os.Stat(searchPath); os.IsNotExist(err) {
			continue
		}

		filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.IsDir() && d.isJavaDirectory(info.Name()) {
				javaExe := filepath.Join(path, "bin", d.javaExeName)
				if _, err := os.Stat(javaExe); err == nil {
					version := d.getJavaVersion(path)
					if version != "" && version != "不明" {
						installations = append(installations, Installation{
							Version: version + " (Directory)",
							Path:    javaExe,
							Home:    path,
							Current: path == d.currentHome,
						})
					}
				}
			}
			return nil
		})
	}

	return installations
}

// Helper functions
func (d *Detector) isValidJavaHome(javaHome string) bool {
	javaExe := filepath.Join(javaHome, "bin", d.javaExeName)
	_, err := os.Stat(javaExe)
	return err == nil
}

func (d *Detector) findAllJavaInPATH() []string {
	var javaExecutables []string

	pathEnv := os.Getenv("PATH")
	pathSeparator := ":"
	if runtime.GOOS == "windows" {
		pathSeparator = ";"
	}

	paths := strings.Split(pathEnv, pathSeparator)
	for _, path := range paths {
		javaPath := filepath.Join(path, d.javaExeName)
		if _, err := os.Stat(javaPath); err == nil {
			javaExecutables = append(javaExecutables, javaPath)
		}
	}

	return javaExecutables
}

func (d *Detector) detectFromHomebrew() []Installation {
	var installations []Installation

	// Common Homebrew Java locations
	brewPaths := []string{
		"/opt/homebrew/opt",
		"/usr/local/opt",
	}

	for _, brewPath := range brewPaths {
		entries, err := os.ReadDir(brewPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() && strings.Contains(strings.ToLower(entry.Name()), "openjdk") {
				javaHome := filepath.Join(brewPath, entry.Name())
				if d.isValidJavaHome(javaHome) {
					version := d.getJavaVersion(javaHome)
					if version != "" && version != "不明" {
						installations = append(installations, Installation{
							Version: version + " (Homebrew)",
							Path:    filepath.Join(javaHome, "bin", d.javaExeName),
							Home:    javaHome,
							Current: javaHome == d.currentHome,
						})
					}
				}
			}
		}
	}

	return installations
}

func (d *Detector) detectFromChocolatey() []Installation {
	var installations []Installation

	chocoPath := filepath.Join("C:", "ProgramData", "chocolatey", "lib")
	if _, err := os.Stat(chocoPath); os.IsNotExist(err) {
		return installations
	}

	entries, err := os.ReadDir(chocoPath)
	if err != nil {
		return installations
	}

	for _, entry := range entries {
		if entry.IsDir() && (strings.Contains(strings.ToLower(entry.Name()), "openjdk") ||
			strings.Contains(strings.ToLower(entry.Name()), "adoptopenjdk")) {
			javaHome := filepath.Join(chocoPath, entry.Name(), "tools")
			if d.isValidJavaHome(javaHome) {
				version := d.getJavaVersion(javaHome)
				if version != "" && version != "不明" {
					installations = append(installations, Installation{
						Version: version + " (Chocolatey)",
						Path:    filepath.Join(javaHome, "bin", d.javaExeName),
						Home:    javaHome,
						Current: javaHome == d.currentHome,
					})
				}
			}
		}
	}

	return installations
}

func (d *Detector) detectFromScoop() []Installation {
	var installations []Installation

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return installations
	}

	scoopPath := filepath.Join(homeDir, "scoop", "apps")
	if _, err := os.Stat(scoopPath); os.IsNotExist(err) {
		return installations
	}

	entries, err := os.ReadDir(scoopPath)
	if err != nil {
		return installations
	}

	for _, entry := range entries {
		if entry.IsDir() && (strings.Contains(strings.ToLower(entry.Name()), "openjdk") ||
			strings.Contains(strings.ToLower(entry.Name()), "adoptopenjdk")) {
			appPath := filepath.Join(scoopPath, entry.Name(), "current")
			if d.isValidJavaHome(appPath) {
				version := d.getJavaVersion(appPath)
				if version != "" && version != "不明" {
					installations = append(installations, Installation{
						Version: version + " (Scoop)",
						Path:    filepath.Join(appPath, "bin", d.javaExeName),
						Home:    appPath,
						Current: appPath == d.currentHome,
					})
				}
			}
		}
	}

	return installations
}

func (d *Detector) removeDuplicates(installations []Installation) []Installation {
	seen := make(map[string]*Installation)
	var result []Installation

	for _, installation := range installations {
		normalizedHome := d.normalizePath(installation.Home)

		if existing, found := seen[normalizedHome]; !found {
			// 初めて見るパス：そのまま追加
			inst := installation
			seen[normalizedHome] = &inst
			result = append(result, inst)
		} else {
			// 重複パス：Currentフラグが立っている方を優先
			if installation.Current && !existing.Current {
				// 新しい方がCurrentなら置き換え
				*existing = installation
				// resultの中身も更新
				for i := range result {
					if d.normalizePath(result[i].Home) == normalizedHome {
						result[i] = installation
						break
					}
				}
			}
		}
	}

	return result
}

// パスを正規化（大文字小文字統一、スラッシュ統一）
func (d *Detector) normalizePath(path string) string {
	// バックスラッシュをスラッシュに統一
	normalized := strings.ReplaceAll(path, "\\", "/")
	// 小文字に統一（Windowsは大文字小文字を区別しない）
	normalized = strings.ToLower(normalized)
	// 末尾のスラッシュを削除
	normalized = strings.TrimSuffix(normalized, "/")
	return normalized
}

// 最終的なCurrentフラグを確定（JAVA_HOMEと一致する1つだけをCurrentにする）
func (d *Detector) ensureSingleCurrent(installations []Installation) []Installation {
	if d.currentHome == "" {
		// JAVA_HOMEが設定されていない場合、全てCurrentをfalseに
		for i := range installations {
			installations[i].Current = false
		}
		return installations
	}

	normalizedCurrentHome := d.normalizePath(d.currentHome)
	foundCurrent := false

	// まず全てのCurrentフラグをfalseにリセット
	for i := range installations {
		installations[i].Current = false
	}

	// JAVA_HOMEと一致する最初の1つだけをCurrentにする
	for i := range installations {
		normalizedHome := d.normalizePath(installations[i].Home)
		if normalizedHome == normalizedCurrentHome {
			installations[i].Current = true
			foundCurrent = true
			break // 最初の1つだけ
		}
	}

	// デバッグ用（オプション）
	if !foundCurrent && d.currentHome != "" {
		// JAVA_HOMEが設定されているのに一致するものがない場合
		// これは正常な状態（検出されなかった別のJavaがJAVA_HOMEに設定されている）
	}

	return installations
}
