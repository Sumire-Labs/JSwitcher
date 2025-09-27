package java

import (
	"os"
)

type Switcher struct{}

func NewSwitcher() *Switcher {
	return &Switcher{}
}

func (s *Switcher) SetJavaHome(javaHome string) error {
	// Set environment variable for current session
	err := os.Setenv("JAVA_HOME", javaHome)
	if err != nil {
		return err
	}

	// On Windows, we should also set it persistently
	// This would require admin privileges, so we'll just set for current session
	// Future enhancement: add persistent setting support

	return nil
}

func (s *Switcher) GetCurrentJavaHome() string {
	return os.Getenv("JAVA_HOME")
}