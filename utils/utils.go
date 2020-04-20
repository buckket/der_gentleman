package utils

import "fmt"

func GenerateTweetURL(username string, id int64) string {
	return fmt.Sprintf("https://twitter.com/%s/status/%d", username, id)
}

func TruncateString(str string, num int) (out string) {
	out = str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		out = str[0:num] + "..."
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
