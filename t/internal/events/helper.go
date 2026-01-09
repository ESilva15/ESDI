package events

import "fmt"

func typeKey(v any) string {
	return fmt.Sprintf("%T", v)
}
