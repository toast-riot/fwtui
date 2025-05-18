package stringsext

func TrimLastChar(s string) string {
	if len(s) > 0 {
		return s[:len(s)-1]
	}
	return s
}
