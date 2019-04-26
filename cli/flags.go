package cli

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

type Flag struct {
	Key         string
	Default     string
	Description string
	Validator   func(string) bool
	Value       *string
}

func ParseFlags(cliFlags []*Flag) (map[string]string, error) {
	parsedValues := make(map[string]string)
	for _, cliFlag := range cliFlags {
		cliFlag.Value = flag.String(cliFlag.Key, cliFlag.Default, cliFlag.Description)
	}
	flag.Parse()
	errorMessages := make([]string, 0)
	for _, cliFlag := range cliFlags {
		if *cliFlag.Value == "" {
			errorMessages = append(errorMessages, fmt.Sprintf("Missing flag \033[32m-%s\033[97m, must be \033[33m%s\033[97m", cliFlag.Key, cliFlag.Description))
		} else if !cliFlag.Validator(*cliFlag.Value) {
			errorMessages = append(errorMessages, fmt.Sprintf("Invalid \033[32m-%s\033[97m: \033[31m%s\033[97m, must be \033[33m%s\033[97m", cliFlag.Key, *cliFlag.Value, cliFlag.Description))
		} else {
			parsedValues[cliFlag.Key] = strings.TrimRight(*cliFlag.Value, "/")
		}
	}

	if len(errorMessages) > 0 {
		return nil, errors.New(fmt.Sprintf("Error parsing CLI flags:\n\t%s", strings.Join(errorMessages, "\n\t")))
	} else {
		return parsedValues, nil
	}
}

func HandleParseFlags(flags map[string]string, err error) (map[string]string, error) {
	if err != nil {
		return nil, err
	} else if flags == nil || len(flags) == 0 {
		return nil, errors.New("Something's gone horribly wrong, parsed CLI flags are empty!")
	}
	return flags, nil
}
