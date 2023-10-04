package machines

import (
	"fmt"
	"strings"
)

func CodeReplace(code []byte, replacements map[string]string) []byte {
	args := make([]string, 0)
	for k, v := range replacements {
		key := fmt.Sprintf("--[[$%s]]--", k)
		args = append(args, key)
		args = append(args, v)
	}

	replacer := strings.NewReplacer(args...)

	return []byte(replacer.Replace(string(code)))
}
