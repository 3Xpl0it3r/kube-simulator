package cluster

import (
	"fmt"
	"sort"
	"strings"
)

func GetArgsList(argsMap map[string]string, extraArgs []string) []string {
	// add extra args to args map to override any default option
	for _, arg := range extraArgs {
		splitArg := strings.SplitN(arg, "=", 2)
		if len(splitArg) < 2 {
			argsMap[splitArg[0]] = "true"
			continue
		}
		argsMap[splitArg[0]] = splitArg[1]
	}
	var args []string
	for arg, value := range argsMap {
		cmd := fmt.Sprintf("--%s=%s", arg, value)
		args = append(args, cmd)
	}
	sort.Strings(args)
	return args
}
