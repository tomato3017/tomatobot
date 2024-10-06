package util

import (
	"fmt"
	"io"

	"golang.org/x/exp/constraints"
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

type NonZeroable interface {
	constraints.Ordered | ~string
}

// FirstNonZero returns the first non-zero value from the provided arguments.
func FirstNonZero[T NonZeroable](values ...T) T {
	var zero T
	for _, value := range values {
		if value != zero {
			return value
		}
	}
	return zero
}
