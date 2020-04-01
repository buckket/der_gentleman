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
