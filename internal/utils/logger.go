package utils

import "log"

func LogInfo(msg string) {
	log.Printf("[INFO] %s", msg)
}

func LogError(msg string, err error) {
	log.Printf("[ERROR] %s: %v", msg, err)
}