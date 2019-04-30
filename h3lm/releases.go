package h3lm

import (
	"k8s.io/helm/pkg/proto/hapi/release"
	"sort"
)

func LatestRelease(releases []*release.Release) *release.Release {
	if releases == nil || len(releases) == 0 {
		return nil
	}

	tmp := make([]*release.Release, len(releases))
	copy(tmp, releases)
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].Info.LastDeployed.Seconds > tmp[j].Info.LastDeployed.Seconds
	})

	return tmp[0]
}

func FilterReleasesByStatusCode(releases []*release.Release, code release.Status_Code) []*release.Release {
	filtered := make([]*release.Release, 0)
	for _, oldRelease := range releases {
		if oldRelease.Info.Status.Code == code {
			filtered = append(filtered, oldRelease)
		}
	}
	return filtered
}
