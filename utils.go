package main

import (
	"fmt"
	"log"
	"time"
)

// Logger colors for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

// LogInfo logs informational messages in blue
func LogInfo(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	log.Printf("%s[%s INFO]%s %s", colorBlue, timestamp, colorReset, message)
}

// LogSuccess logs success messages in green
func LogSuccess(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	log.Printf("%s[%s SUCCESS]%s %s", colorGreen, timestamp, colorReset, message)
}

// LogWarning logs warning messages in yellow
func LogWarning(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	log.Printf("%s[%s WARNING]%s %s", colorYellow, timestamp, colorReset, message)
}

// LogError logs error messages in red
func LogError(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	log.Printf("%s[%s ERROR]%s %s", colorRed, timestamp, colorReset, message)
}

// LogDebug logs debug messages in cyan
func LogDebug(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	log.Printf("%s[%s DEBUG]%s %s", colorCyan, timestamp, colorReset, message)
}

// Contains checks if a string slice contains a specific string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
