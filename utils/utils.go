package utils

import (
	"fmt"
	"unicode/utf8"
)

func GenerateTweetURL(username string, id int64) string {
	return fmt.Sprintf("https://twitter.com/%s/status/%d", username, id)
}

func TruncateString(str string, num int) (out string) {
	out = str
	chars := 0
	for i, r := range out {
		if chars >= num-2 {
			out = out[:i] + "…"
			break
		}
		rlen := utf8.RuneLen(r)
		if rlen > 1 {
			chars += 2
		} else {
			chars++
		}
	}
	return out
}

func TopLikers(top interface{}) []string {
	nested, ok := top.([]interface{})
	if ok {
		var out []string
		for _, i := range nested {
			switch s := i.(type) {
			case string:
				out = append(out, s)
			case []string:
				out = append(out, s...)
			}
		}
		return out
	} else {
		switch s := top.(type) {
		case string:
			return []string{s}
		case []string:
			return s
		}
		return nil
	}
}
