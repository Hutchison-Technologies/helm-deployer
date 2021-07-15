package h3lm

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/release"
	"testing"
)

func makeReleaseWithCode(code release.Status_Code) *release.Release {
	return &release.Release{
		Info: &release.Info{
			Status: &release.Status{
				Code: code,
			},
		},
	}
}

func makeReleaseWithTime(epoch int64) *release.Release {
	return &release.Release{
		Info: &release.Info{
			LastDeployed: &timestamp.Timestamp{
				Seconds: epoch,
			},
		},
	}
}

func Test_LatestRelease_Returns_Nil_When_Given_Nil(t *testing.T) {
	assert.Nil(t, LatestRelease(nil))
}

func Test_LatestRelease_Returns_Nil_When_Given_Empty_Array(t *testing.T) {
	assert.Nil(t, LatestRelease(make([]*release.Release, 0)))
}

func Test_LatestRelease_Returns_First_Element_When_Given_Array_Of_One(t *testing.T) {
	first := makeReleaseWithTime(100)
	assert.Equal(t, first, LatestRelease([]*release.Release{
		first,
	}))
}

func Test_LatestRelease_Returns_Latest_Release(t *testing.T) {
	latest := makeReleaseWithTime(100)
	assert.Equal(t, latest, LatestRelease([]*release.Release{
		makeReleaseWithTime(80),
		makeReleaseWithTime(85),
		latest,
		makeReleaseWithTime(87),
		makeReleaseWithTime(90),
	}))
}

func Test_LatestRelease_Does_Not_Mutate_Input_Array(t *testing.T) {
	latest := makeReleaseWithTime(100)
	input := []*release.Release{
		makeReleaseWithTime(80),
		makeReleaseWithTime(85),
		latest,
		makeReleaseWithTime(87),
		makeReleaseWithTime(90),
	}
	inputCopy := make([]*release.Release, len(input))
	copy(inputCopy, input)
	LatestRelease(input)
	assert.Equal(t, inputCopy, input)
}

func Test_FilterReleasesByStatusCode_Returns_Empty_Array_When_Given_Nil(t *testing.T) {
	assert.Equal(t, make([]*release.Release, 0), FilterReleasesByStatusCode(nil, release.Status_DEPLOYED))
}

func Test_FilterReleasesByStatusCode_Returns_Empty_Array_When_Given_Empty_Array(t *testing.T) {
	assert.Equal(t, make([]*release.Release, 0), FilterReleasesByStatusCode(make([]*release.Release, 0), release.Status_DEPLOYED))
}

func Test_FilterReleasesByStatusCode_Returns_Empty_Array_When_None_Match(t *testing.T) {
	input := []*release.Release{
		makeReleaseWithCode(release.Status_DELETED),
		makeReleaseWithCode(release.Status_DELETING),
		makeReleaseWithCode(release.Status_FAILED),
	}
	assert.Equal(t, make([]*release.Release, 0), FilterReleasesByStatusCode(input, release.Status_DEPLOYED))
}

func Test_FilterReleasesByStatusCode_Returns_Entire_Array_When_All_Match(t *testing.T) {
	code := release.Status_DEPLOYED
	input := []*release.Release{
		makeReleaseWithCode(code),
		makeReleaseWithCode(code),
		makeReleaseWithCode(code),
	}
	assert.Equal(t, input, FilterReleasesByStatusCode(input, code))
}

func Test_FilterReleasesByStatusCode_Returns_Partial_Array_When_Some_Match(t *testing.T) {
	code := release.Status_DEPLOYED
	input := []*release.Release{
		makeReleaseWithCode(code),
		makeReleaseWithCode(code),
		makeReleaseWithCode(code),
		makeReleaseWithCode(release.Status_DELETED),
		makeReleaseWithCode(release.Status_DELETING),
		makeReleaseWithCode(release.Status_FAILED),
	}
	assert.Equal(t, []*release.Release{
		makeReleaseWithCode(code),
		makeReleaseWithCode(code),
		makeReleaseWithCode(code),
	}, FilterReleasesByStatusCode(input, code))
}
