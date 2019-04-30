package cli

import (
	"fmt"
	"k8s.io/helm/pkg/proto/hapi/release"
	"log"
	"time"
)

func PrintMap(m map[string]string) {
	for key, value := range m {
		log.Printf("\t%s: %s", key, Green(value))
	}
}

func PrintRelease(rel *release.Release) {
	log.Printf("\n\tName: %s\n\tRelease Version: %s\n\tStatus: %s\n\tLast Deployed: %s",
		Green(rel.Name), Green(string(rel.Version)), Green(rel.Info.Status.Code.String()), Green(time.Unix(rel.Info.LastDeployed.Seconds, int64(rel.Info.LastDeployed.Nanos)).String()))
}

//TODO fix this

func Green(str string) string {
	return fmt.Sprintf("\033[32m%s\033[97m", str)
}

func Orange(str string) string {
	return fmt.Sprintf("\033[33m%s\033[97m", str)
}
