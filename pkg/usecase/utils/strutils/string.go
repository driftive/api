package strutils

func OrEmpty(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

func OrNil(str string) *string {
	if str == "" {
		return nil
	}
	return &str
}
