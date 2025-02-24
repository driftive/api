package parsing

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

func PInt64ToString(i *int64) string {
	return strconv.FormatInt(*i, 10)
}
