package utils

import "strings"

func TrimPrefix(msg string, prefix string, mention string) (isCmd bool, trimmedMsg string) {
	isCmd = false
	if strings.HasPrefix(msg, prefix) {
		isCmd = true
		trimmedMsg = strings.TrimPrefix(msg, prefix)
	} else if strings.HasPrefix(msg, mention) {
		isCmd = true
		trimmedMsg = strings.TrimPrefix(msg, mention)
		trimmedMsg = strings.TrimPrefix(trimmedMsg, " ")
	}
	return
}
