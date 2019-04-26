package cli

import (
	"k8s.io/helm/pkg/proto/hapi/release"
	"log"
	"time"
)

func PrintMap(m map[string]string) {
	for key, value := range m {
		log.Printf("\t%s: \033[32m%s\033[97m", key, value)
	}
}

func PrintRelease(rel *release.Release) {
	log.Printf("\n\tName: \033[32m%s\033[97m\n\tRelease Number: \033[32m%d\033[97m\n\tStatus: \033[32m%s\033[97m\n\tLast Deployed: \033[32m%s\033[97m",
		rel.Name, rel.Version, rel.Info.Status.Code.String(), time.Unix(rel.Info.LastDeployed.Seconds, int64(rel.Info.LastDeployed.Nanos)).String())
}
