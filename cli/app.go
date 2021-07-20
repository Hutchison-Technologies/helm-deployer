package cli

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"bytes"
	"bufio"
	goYaml "github.com/ghodss/yaml"

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

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/kube"
)

const (
	CHART_DIR               = "chart-dir"
	APP_NAME                = "app-name"
	APP_VERSION             = "app-version"
	TARGET_ENV              = "target-env"
	DEFAULT_COLOUR          = "blue"
	RELEASE_UPGRADE_TIMEOUT = 900
	ROLLBACK_VERSION_POOL   = 50
	ROLLBACK_TIMEOUT        = 900
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

func buildHelmConfig() *action.Configuration {
    log.Println("Building helm configuration..")

	kube.ManagedFieldsManager = "helm"
	helmConfig := new(action.Configuration) 
	settings := cli.New()

	if err := helmConfig.Init(settings.RESTClientGetter(), "default",
		os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}

	log.Printf("Configured helm configuration.")
    return helmConfig
}

func releaseWithValues(releaseName string, chartValuesYaml *yaml.Yaml, chartValuesEdits [][]interface{}, helmConfig *action.Configuration, chartDir string) *release.Release {
	log.Printf("Editing chart values to deploy %s..", Green(releaseName))
	chartValues := editChartValues(chartValuesYaml, chartValuesEdits)
	log.Printf("Successfully edited chart values:\n%s", Orange(string(chartValues)))

	log.Printf("Deploying: %s..", Green(releaseName))
	deployedRelease, err := deployRelease(helmConfig, releaseName, chartDir, chartValues)
	if err != nil {
		log.Printf("Error deploying %s: %s", Green(releaseName), err.Error())
		log.Println("Determining whether rollback is necessary..")
		if shouldRollBack(helmConfig, releaseName) {
			log.Println("Rollback is necessary")
			rollbackErr := rollback(helmConfig, releaseName)
			runtime.PanicIfError(rollbackErr)
		} else {
			log.Println("Current release is ok, nothing to do")
		}
		panic(fmt.Errorf("Original deploy error: %s", err))
	} else {
		return deployedRelease
	}
}


func deployRelease(helmConfig *action.Configuration, releaseName, chartDir string, chartValues []byte) (*release.Release, error) {
	log.Printf("Checking for existing %s release..", Green(releaseName))

	releaseNamespace := "default"

	// Create new fetchManager to get information about existing releases
	fetchManager := action.NewGet(helmConfig)
	releaseContent, err := fetchManager.Run(releaseName)
	existingReleaseCode := release.StatusUnknown

	if releaseContent != nil {
		log.Println("Found existing release:")
		PrintRelease(releaseContent)
		existingReleaseCode = releaseContent.Info.Status
	}

	switch deployment.DetermineReleaseCourse(releaseName, existingReleaseCode, err) {
	case deployment.ReleaseCourse.INSTALL:
		log.Println("No existing release found, installing release..")
		
		chart, err := loader.Load(chartDir)
		if err != nil {
			panic(err)
		}
		log.Println("Chart: ", chart)

		duration, err := time.ParseDuration("300s")
		if err != nil {
			panic(err)
		}

		installManager := action.NewInstall(helmConfig)
		installManager.Namespace = releaseNamespace
		installManager.ReleaseName = releaseName
		installManager.Wait = true
		installManager.WaitForJobs = true
		installManager.Timeout = duration
		installManager.Description = "Some chart"
		// Disable when not testing...
		installManager.DryRun = false


		vals := make(map[string]interface{})
		err = goYaml.Unmarshal(chartValues, &vals)
		if err != nil {
			panic(err)
		}

		// Push values to chart and install
		installResponse, err := installManager.Run(chart, vals)
		if err != nil {
			return nil, err
		}
		log.Println("Installed release: ", installResponse)
		return installResponse, nil
	case deployment.ReleaseCourse.UPGRADE_WITH_DIFF_CHECK:
		log.Println("Dry-running release to obtain full manifest..")

		dryRunRelease, err := upgradeRelease(helmConfig, releaseName, chartDir, chartValues, true)
		if err != nil {
			return nil, err
		}

		currentManifestString := string(releaseContent.Manifest)
		newManifestString := string(dryRunRelease.Manifest)

		currentManifests := manifest.Parse(currentManifestString, "default")
		newManifests := manifest.Parse(newManifestString, "default")

		log.Println("Checking proposed release for changes against existing release..")
		
		var b bytes.Buffer
		var manifestContext int = -1

		hasChanges := diff.Manifests(currentManifests, newManifests, []string{}, false, manifestContext, bufio.NewWriter(&b))
		if !hasChanges {
			return nil, errors.New("No difference detected between this release and the existing release, no deploy.")
		}
		fallthrough
	case deployment.ReleaseCourse.UPGRADE:
		log.Printf("Upgrading release, will timeout after %d seconds..", RELEASE_UPGRADE_TIMEOUT)
		upgradeRelease, err := upgradeRelease(helmConfig, releaseName, chartDir, chartValues, false)
		if err != nil {
			return nil, err
		}
		return upgradeRelease, nil
	}

	return nil, errors.New("Unknown release course")
}


func upgradeRelease(helmConfig *action.Configuration, releaseName, chartDir string, chartValues []byte, dryRun bool) (*release.Release, error) {
	chart, err := loader.Load(chartDir)
    if err != nil {
        panic(err)
    }
	log.Println("Chart: ", chart)

	duration, err := time.ParseDuration("300s")
	if err != nil {
		panic(err)
	}

	upgradeManager := action.NewUpgrade(helmConfig)
	upgradeManager.Force = true;
	upgradeManager.Recreate = true;
	upgradeManager.Wait = true;
	upgradeManager.Timeout = duration
	// Remove when finished testing
	upgradeManager.DryRun = dryRun

	vals := make(map[string]interface{})
	err = goYaml.Unmarshal(chartValues, &vals)
	if err != nil {
		panic(err)
	}

	// Push values to upgrade request
	res, err := upgradeManager.Run(releaseName, chart, vals)

	if err != nil {
		return nil, err
	}
	return res, nil
}


func shouldRollBack(helmConfig *action.Configuration, releaseName string) bool {
	status := action.NewStatus(helmConfig)
	releaseStatus, err := status.Run(releaseName)
	runtime.PanicIfError(err)
	return releaseStatus.Info.Status != release.StatusDeployed
}


func rollback(helmConfig *action.Configuration, releaseName string) error {
	log.Printf("Gathering up to the last %d release(s)..", ROLLBACK_VERSION_POOL)

	status := action.NewHistory(helmConfig)
	status.Max = ROLLBACK_VERSION_POOL

	releaseHistory, err := status.Run(releaseName)

	runtime.PanicIfError(err)
	if len(releaseHistory) == 0 {
		return errors.New("No prior release(s) to roll back to!")
	}

	log.Printf("Found %d prior release(s), filtering for successful release(s)..", len(releaseHistory))
	successfullyDeployedReleases := h3lm.FilterReleasesByStatusCode(releaseHistory, release.StatusDeployed)

	if len(successfullyDeployedReleases) == 0 {
		return errors.New("No successfully deployed prior release(s) to roll back to!")
	}

	log.Printf("Found %d prior successful release(s), finding the latest..", len(successfullyDeployedReleases))
	latestSuccessfulRelease := h3lm.LatestRelease(successfullyDeployedReleases)

	log.Println("Latest successful release:")
	PrintRelease(latestSuccessfulRelease)

	log.Println("Rolling back..")

	duration, err := time.ParseDuration(fmt.Sprintf("%ds", ROLLBACK_TIMEOUT))
	if err != nil {
		panic(err)
	}

	rollbackManager := action.NewRollback(helmConfig)
	rollbackManager.Force = true;
	rollbackManager.Recreate = true;
	rollbackManager.Wait = true;
	rollbackManager.Timeout = duration
	rollbackManager.Version = latestSuccessfulRelease.Version;

	err = rollbackManager.Run(releaseName)

	if err != nil {
		return fmt.Errorf("Failed to rollback: %s", err)
	}
	log.Printf("Successfully rolled %s back:", Green(releaseName))
	return nil
}

func deleteHPA(offlineHPAName string) error {
	hpaClient := kubeCtlHPAClient().HorizontalPodAutoscalers(apiv1.NamespaceDefault)
	deletePolicy := metav1.DeletePropagationBackground
	deletionError := hpaClient.Delete(context.TODO(), offlineHPAName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if deletionError != nil {
		log.Printf("Error deleting HPA (%s): %v", hpaClient, deletionError)
		return deletionError
	}
	log.Printf("Success! Removed the HPA (%s).", offlineHPAName)
	return nil
}

func scaleReplicaSet(offlineDeploymentName string, scaleSize int32) error {
	deploymentsClient := kubeCtlAppClient().Deployments(apiv1.NamespaceDefault)
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := deploymentsClient.Get(context.TODO(), offlineDeploymentName, metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("Failed to get latest version of Deployment: %v", getErr))
		}

		var numberOfReplicas int32 = scaleSize
		result.Spec.Replicas = &numberOfReplicas

		_, updateErr := deploymentsClient.Update(context.TODO(), result, metav1.UpdateOptions{})
		return updateErr
	})

	return retryErr
}
