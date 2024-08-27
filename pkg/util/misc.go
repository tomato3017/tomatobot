package util

import (
	"fmt"
	"io"
)

func CloseSafely(closer io.Closer) {
	if closer == nil {
		return
	}

	err := closer.Close()
	if err != nil {
		fmt.Printf("Error closing resource: %v\n", err)
	}
}

func TruncateString(str string, length int, appendStr string) string {
	if len(str) <= length {
		return str
	}

	return fmt.Sprintf("%s%s", str[:length], "...")
}
