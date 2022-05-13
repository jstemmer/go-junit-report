package pass

import "fmt"

func message() string {
	return "line"
}

func unused(n int) string {
	if n < 0 {
		return "0"
	}
	return fmt.Sprintf("%d\n", n)
}
