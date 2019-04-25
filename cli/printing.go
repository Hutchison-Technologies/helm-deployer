package cli

import (
	"log"
)

func PrintMap(m map[string]string) {
	for key, value := range m {
		log.Printf("\t%s: \033[32m%s\033[97m", key, value)
	}
}
