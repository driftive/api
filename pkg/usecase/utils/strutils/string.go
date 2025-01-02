package strutils

import "strconv"

func StringToInt64(str string) int64 {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		panic(err)
	}
	return i
}

func Int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}
