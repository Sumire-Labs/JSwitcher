package java

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	searchPaths  []string
	javaExeName  string
	currentHome  string
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

	// Check PATH for java executable
	if javaPath, err := exec.LookPath("java"); err == nil {
		if javaHome := d.extractJavaHomeFromExe(javaPath); javaHome != "" {
			version := d.getJavaVersion(javaHome)
			installations = append(installations, Installation{
				Version: version,
				Path:    javaPath,
				Home:    javaHome,
				Current: javaHome == d.currentHome,
			})
		}
	}

	// Search common installation directories
	for _, searchPath := range d.searchPaths {
		if _, err := os.Stat(searchPath); os.IsNotExist(err) {
			continue
		}

		filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			// Look for directories that contain Java installations
			if info.IsDir() && d.isJavaDirectory(info.Name()) {
				javaExe := filepath.Join(path, "bin", d.javaExeName)
				if _, err := os.Stat(javaExe); err == nil {
					version := d.getJavaVersion(path)
					if version != "" && version != "不明" {
						// Check if already added
						if !d.containsInstallation(installations, path) {
							installations = append(installations, Installation{
								Version: version,
								Path:    javaExe,
								Home:    path,
								Current: path == d.currentHome,
							})
						}
					}
				}
			}
			return nil
		})
	}

	// Sort installations
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