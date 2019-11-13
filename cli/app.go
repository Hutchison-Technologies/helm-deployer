package cli

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Hutchison-Technologies/helm-deployer/charts"
	"github.com/Hutchison-Technologies/helm-deployer/deployment"
	"github.com/Hutchison-Technologies/helm-deployer/gosexy/yaml"
	"github.com/Hutchison-Technologies/helm-deployer/h3lm"
	"github.com/Hutchison-Technologies/helm-deployer/kubectl"
	"github.com/Hutchison-Technologies/helm-deployer/runtime"
	"github.com/databus23/helm-diff/diff"
	"github.com/databus23/helm-diff/manifest"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	autoscalingv1 "k8s.io/client-go/kubernetes/typed/autoscaling/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/util/retry"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/proto/hapi/services"
)

const (
	CHART_DIR               = "chart-dir"
	APP_NAME                = "app-name"
	APP_VERSION             = "app-version"
	TARGET_ENV              = "target-env"
	DEFAULT_COLOUR          = "blue"
	RELEASE_UPGRADE_TIMEOUT = 300
	ROLLBACK_VERSION_POOL   = 50
	ROLLBACK_TIMEOUT        = 600
)

func Run() error {
	log.Println("Starting helm-deployer..")

	switch DetermineCommand(os.Args[1]) {
	case Command.BLUEGREEN:
		log.Println("Running bluegreen deploy..")
		return RunBlueGreenDeploy()
	case Command.STANDARD_CHART:
		log.Println("Running standard-chart deploy..")
		return RunStandardChartDeploy()
	case Command.MICROSERVICE:
		log.Println("Running microservice deploy..")
		return RunMicroserviceDeploy()
	default:
		return errors.New(fmt.Sprintf("Unknown command: %s\nShould be one of: %s", Green(os.Args[1]), strings.Join([]string{Orange(Command.BLUEGREEN), Orange(Command.STANDARD_CHART)}, ", ")))
	}
}

func parseCLIFlags(flagsToParse []*Flag) map[string]string {
	potentialParsedFlags, potentialParseFlagsErr := ParseFlags(flagsToParse)
	cliFlags, err := HandleParseFlags(potentialParsedFlags, potentialParseFlagsErr)
	runtime.PanicIfError(err)
	return cliFlags
}

func loadChartValues(chartDir, targetEnv string) *yaml.Yaml {
	chartValuesPath := deployment.ChartValuesPath(chartDir, targetEnv)
	values, err := charts.LoadValuesYaml(chartValuesPath)
	runtime.PanicIfError(err)
	return values
}

func editChartValues(valuesYaml *yaml.Yaml, settings [][]interface{}) []byte {
	values, err := charts.EditValuesYaml(valuesYaml, settings)
	runtime.PanicIfError(err)
	return values
}

func kubeCtlClient() v1.CoreV1Interface {
	client, err := kubectl.Client()
	runtime.PanicIfError(err)
	return client
}

func kubeCtlAppClient() appv1.AppsV1Interface {
	client, err := kubectl.AppsClient()
	runtime.PanicIfError(err)
	return client
}

func kubeCtlHPAClient() autoscalingv1.AutoscalingV1Interface {
	client, err := kubectl.HPAClient()
	runtime.PanicIfError(err)
	return client
}

func buildHelmClient() *helm.Client {
	log.Println("Setting up tiller tunnel..")
	tillerHost, err := kubectl.SetupTillerTunnel()
	runtime.PanicIfError(err)
	log.Println("Established tiller tunnel")
	helmClient := helm.NewClient(helm.Host(tillerHost), helm.ConnectTimeout(60))
	log.Printf("Configured helm client, pinging tiller at: %s..", Green(tillerHost))
	err = helmClient.PingTiller()
	runtime.PanicIfError(err)
	return helmClient
}

func releaseWithValues(releaseName string, chartValuesYaml *yaml.Yaml, chartValuesEdits [][]interface{}, helmClient *helm.Client, chartDir string) *release.Release {
	log.Printf("Editing chart values to deploy %s..", Green(releaseName))
	chartValues := editChartValues(chartValuesYaml, chartValuesEdits)
	log.Printf("Successfully edited chart values:\n%s", Orange(string(chartValues)))

	log.Printf("Deploying: %s..", Green(releaseName))
	deployedRelease, err := deployRelease(helmClient, releaseName, chartDir, chartValues)
	if err != nil {
		log.Printf("Error deploying %s: %s", Green(releaseName), err.Error())
		log.Println("Determining whether rollback is necessary..")
		if shouldRollBack(helmClient, releaseName) {
			log.Println("Rollback is necessary")
			rollbackErr := rollback(helmClient, releaseName)
			runtime.PanicIfError(rollbackErr)
		} else {
			log.Println("Current release is ok, nothing to do")
		}
		panic(fmt.Errorf("Original deploy error: %s", err))
	} else {
		return deployedRelease
	}
}

func deployRelease(helmClient *helm.Client, releaseName, chartDir string, chartValues []byte) (*release.Release, error) {
	log.Printf("Checking for existing %s release..", Green(releaseName))
	releaseContent, err := helmClient.ReleaseContent(releaseName)
	existingReleaseCode := release.Status_UNKNOWN

	if releaseContent != nil {
		log.Println("Found existing release:")
		PrintRelease(releaseContent.Release)
		existingReleaseCode = releaseContent.Release.Info.Status.Code
	}

	switch deployment.DetermineReleaseCourse(releaseName, existingReleaseCode, err) {
	case deployment.ReleaseCourse.INSTALL:
		log.Println("No existing release found, installing release..")
		installResponse, err := helmClient.InstallRelease(chartDir, "default", helm.ReleaseName(releaseName), helm.InstallWait(true), helm.InstallTimeout(300), helm.InstallDescription("Some chart"), helm.ValueOverrides(chartValues))
		if err != nil {
			return nil, err
		}
		return installResponse.Release, nil

	case deployment.ReleaseCourse.UPGRADE_WITH_DIFF_CHECK:
		log.Println("Dry-running release to obtain full manifest..")

		dryRunResponse, err := upgradeRelease(helmClient, releaseName, chartDir, chartValues, helm.UpgradeDryRun(true))
		if err != nil {
			return nil, err
		}

		currentManifests := manifest.ParseRelease(releaseContent.Release)
		newManifests := manifest.ParseRelease(dryRunResponse.Release)

		log.Println("Checking proposed release for changes against existing release..")
		hasChanges := diff.DiffManifests(currentManifests, newManifests, []string{}, -1, os.Stdout)
		if !hasChanges {
			return nil, errors.New("No difference detected between this release and the existing release, no deploy.")
		}
		fallthrough
	case deployment.ReleaseCourse.UPGRADE:
		log.Printf("Upgrading release, will timeout after %d seconds..", RELEASE_UPGRADE_TIMEOUT)
		upgradeResponse, err := upgradeRelease(helmClient, releaseName, chartDir, chartValues)
		if err != nil {
			return nil, err
		}
		return upgradeResponse.Release, nil
	}

	return nil, errors.New("Unknown release course")
}

func upgradeRelease(helmClient *helm.Client, releaseName, chartDir string, chartValues []byte, opts ...helm.UpdateOption) (*services.UpdateReleaseResponse, error) {
	opts = append(opts, helm.UpgradeForce(true), helm.UpgradeRecreate(true), helm.UpgradeWait(true), helm.UpgradeTimeout(300), helm.UpdateValueOverrides(chartValues))
	res, err := helmClient.UpdateRelease(releaseName, chartDir, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func shouldRollBack(helmClient *helm.Client, releaseName string) bool {
	releaseStatus, err := helmClient.ReleaseStatus(releaseName)
	runtime.PanicIfError(err)
	return releaseStatus.Info.Status.Code != release.Status_DEPLOYED
}

func rollback(helmClient *helm.Client, releaseName string) error {
	log.Printf("Gathering up to the last %d release(s)..", ROLLBACK_VERSION_POOL)
	releaseHistory, err := helmClient.ReleaseHistory(releaseName, helm.WithMaxHistory(ROLLBACK_VERSION_POOL))
	runtime.PanicIfError(err)
	if len(releaseHistory.Releases) == 0 {
		return errors.New("No prior release(s) to roll back to!")
	}

	log.Printf("Found %d prior release(s), filtering for successful release(s)..", len(releaseHistory.Releases))
	successfullyDeployedReleases := h3lm.FilterReleasesByStatusCode(releaseHistory.Releases, release.Status_DEPLOYED)

	if len(successfullyDeployedReleases) == 0 {
		return errors.New("No successfully deployed prior release(s) to roll back to!")
	}

	log.Printf("Found %d prior successful release(s), finding the latest..", len(successfullyDeployedReleases))
	latestSuccessfulRelease := h3lm.LatestRelease(successfullyDeployedReleases)

	log.Println("Latest successful release:")
	PrintRelease(latestSuccessfulRelease)

	log.Println("Rolling back..")
	rollbackResponse, err := helmClient.RollbackRelease(releaseName, helm.RollbackForce(true), helm.RollbackRecreate(true), helm.RollbackWait(true), helm.RollbackTimeout(ROLLBACK_TIMEOUT), helm.RollbackVersion(latestSuccessfulRelease.Version))
	if err != nil {
		return fmt.Errorf("Failed to rollback: %s", err)
	}
	log.Printf("Successfully rolled %s back:", Green(releaseName))
	PrintRelease(rollbackResponse.Release)
	return nil
}

func deleteHPA(offlineHPAName string) error {
	hpaClient := kubeCtlHPAClient().HorizontalPodAutoscalers(apiv1.NamespaceDefault)
	deletePolicy := metav1.DeletePropagationBackground
	deletionError := hpaClient.Delete(offlineHPAName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if deletionError != nil {
		log.Fatalf("Error deleting HPA (%s): %v", hpaClient, deletionError)
	}
	log.Printf("Success! Removed the HPA (%s).", offlineHPAName)
	return nil
}

func scaleReplicaSet(offlineDeploymentName string, scaleSize int32) error {
	deploymentsClient := kubeCtlAppClient().Deployments(apiv1.NamespaceDefault)
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := deploymentsClient.Get(offlineDeploymentName, metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("Failed to get latest version of Deployment: %v", getErr))
		}

		var numberOfReplicas int32 = scaleSize
		result.Spec.Replicas = &numberOfReplicas

		_, updateErr := deploymentsClient.Update(result)
		return updateErr
	})

	return retryErr
}
