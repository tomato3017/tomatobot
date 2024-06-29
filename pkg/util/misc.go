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
