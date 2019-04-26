package cli

import (
	"errors"
	"fmt"
	"github.com/Hutchison-Technologies/bluegreen-deployer/deployment"
	"github.com/Hutchison-Technologies/bluegreen-deployer/filesystem"
	"github.com/Hutchison-Technologies/bluegreen-deployer/k8s"
	"github.com/Hutchison-Technologies/bluegreen-deployer/runtime"
	"github.com/databus23/helm-diff/diff"
	"github.com/databus23/helm-diff/manifest"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/portforwarder"
	"k8s.io/helm/pkg/proto/hapi/release"
	storageerrors "k8s.io/helm/pkg/storage/errors"
	"log"
	"os"
	"strings"
)

const (
	CHART_DIR   = "chart-dir"
	APP_NAME    = "app-name"
	APP_VERSION = "app-version"
	TARGET_ENV  = "target-env"
)

var flags = []*Flag{
	&Flag{
		Key:         CHART_DIR,
		Default:     "./chart",
		Description: "directory containing the service-to-be-deployed's chart definition.",
		Validator:   filesystem.IsDirectory,
	},
	&Flag{
		Key:         APP_NAME,
		Default:     "",
		Description: "name of the service-to-be-deployed (lower-case, alphanumeric + dashes).",
		Validator:   deployment.IsValidAppName,
	},
	&Flag{
		Key:         APP_VERSION,
		Default:     "",
		Description: "semantic version of the service-to-be-deployed (vX.X.X, or X.X.X).",
		Validator:   deployment.IsValidAppVersion,
	},
	&Flag{
		Key:         TARGET_ENV,
		Default:     "",
		Description: "name of the environment in which to deploy the service (prod or staging).",
		Validator:   deployment.IsValidTargetEnv,
	},
}

func parseCLIFlags(flagsToParse []*Flag) map[string]string {
	potentialParsedFlags, potentialParseFlagsErr := ParseFlags(flagsToParse)
	cliFlags, err := HandleParseFlags(potentialParsedFlags, potentialParseFlagsErr)
	runtime.PanicIfError(err)
	return cliFlags
}

func Run() error {
	log.Println("Starting bluegreen-deployer..")

	log.Println("Parsing CLI flags..")
	cliFlags := parseCLIFlags(flags)
	log.Println("Successfully parsed CLI flags:")
	PrintMap(cliFlags)

	// cliFlags[APP_VERSION]

	log.Println("Locating chart values..")
	chartValues, chartValuesErr := locateChartValues(cliFlags[CHART_DIR], cliFlags[TARGET_ENV])
	if chartValuesErr != nil {
		panic(chartValuesErr.Error())
	}
	log.Println("Successfully located chart values")

	log.Println("Accessing kubernetes..")
	kubernetes := kubernetesCoreV1()
	log.Println("Determining deploy colour..")
	deployColour := determineDeployColour(cliFlags[TARGET_ENV], cliFlags[APP_NAME], kubernetes)
	log.Printf("Determined deploy colour: \033[32m%s\033[97m", deployColour)
	deploymentName := deployment.DeploymentName(cliFlags[TARGET_ENV], deployColour, cliFlags[APP_NAME])
	helmClient := buildHelmClient()
	log.Printf("Deploying: \033[32m%s\033[97m..", deploymentName)

	log.Printf("Loading values from: \033[32m%s\033[97m..", chartValues)
	values, valueReadErr := ioutil.ReadFile(chartValues)
	if valueReadErr != nil {
		panic(valueReadErr.Error())
	}

	log.Printf("Checking for existing \033[32m%s\033[97m release..", deploymentName)
	releaseContent, err := helmClient.ReleaseContent(deploymentName)
	if err != nil && strings.Contains(err.Error(), storageerrors.ErrReleaseNotFound(deploymentName).Error()) {
		log.Println("No existing release found, installing chart..")
		res, installErr := helmClient.InstallRelease(cliFlags[CHART_DIR], "default", helm.ReleaseName(deploymentName), helm.InstallWait(true), helm.InstallTimeout(300), helm.InstallDescription("Some chart"), helm.ValueOverrides(values))
		if installErr != nil {
			panic(installErr.Error())
		}
		log.Printf("Successfully installed: %s", res.Release.Info.Description)
	} else if releaseContent.Release.Info.Status.Code == release.Status_DELETED {
		log.Println("Existing release found in DELETED state, upgrading chart..")
		res, updateErr := helmClient.UpdateRelease(deploymentName, cliFlags[CHART_DIR], helm.UpgradeForce(true), helm.UpgradeRecreate(true), helm.UpgradeWait(true), helm.UpgradeTimeout(300), helm.UpdateValueOverrides(values))
		if updateErr != nil {
			panic(updateErr.Error())
		}
		log.Printf("Successfully upgraded: %s", res.Release.Info.Description)
	} else {
		log.Println("Existing release found, performing dry-run release..")

		dryRunResponse, dryRunErr := helmClient.UpdateRelease(deploymentName, cliFlags[CHART_DIR], helm.UpgradeDryRun(true), helm.UpgradeForce(true), helm.UpgradeRecreate(true), helm.UpgradeWait(true), helm.UpgradeTimeout(300), helm.UpdateValueOverrides(values))
		if dryRunErr != nil {
			panic(dryRunErr.Error())
		}

		currentManifests := manifest.ParseRelease(releaseContent.Release)
		newManifests := manifest.ParseRelease(dryRunResponse.Release)

		log.Println("Checking proposed release for changes against existing release..")
		hasChanges := diff.DiffManifests(currentManifests, newManifests, []string{}, -1, os.Stderr)
		if hasChanges {
			res, updateErr := helmClient.UpdateRelease(deploymentName, cliFlags[CHART_DIR], helm.UpgradeForce(true), helm.UpgradeRecreate(true), helm.UpgradeWait(true), helm.UpgradeTimeout(300), helm.UpdateValueOverrides(values))
			if updateErr != nil {
				panic(updateErr.Error())
			}
			log.Printf("Successfully upgraded: %s", res.Release.Info.Description)
		} else {
			log.Println("No difference detected, no deploy.")
		}
	}
	return nil
}

func locateChartValues(chartDir, targetEnv string) (string, error) {
	chartValues := deployment.ChartValuesPath(chartDir, targetEnv)
	if !filesystem.IsFile(chartValues) {
		return "", errors.New(fmt.Sprintf("Expected to find chart values yaml at: \033[31m%s\033[97m, but found nothing.", chartValues))
	} else {
		return chartValues, nil
	}
}

func kubernetesCoreV1() v1.CoreV1Interface {
	_, client, err := getKubeClient()
	if err != nil {
		panic(err.Error())
	}
	log.Println("Successfully created kubectl client")
	return client.CoreV1()
}

func buildHelmClient() *helm.Client {
	log.Println("Setting up tiller tunnel..")
	tillerHost, err := setupTillerTunnel()
	if err != nil {
		panic(err.Error())
	}
	log.Println("Established tiller tunnel")
	helmClient := helm.NewClient(helm.Host(tillerHost), helm.ConnectTimeout(60))
	log.Printf("Configured helm client, pinging tiller at: \033[32m%s\033[97m", tillerHost)
	err = helmClient.PingTiller()
	if err != nil {
		panic(err.Error())
	}
	log.Println("Successfully initialised helm client!")
	return helmClient
}

func setupTillerTunnel() (string, error) {
	config, client, err := getKubeClient()
	if err != nil {
		return "", err
	}

	tillerTunnel, err := portforwarder.New("kube-system", client, config)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("127.0.0.1:%d", tillerTunnel.Local), nil
}

// getKubeClient creates a Kubernetes config and client for a given kubeconfig context.
func getKubeClient() (*rest.Config, kubernetes.Interface, error) {
	homeDir := os.Getenv("HOME")
	log.Printf("Using home dir: \033[32m%s\033[97m", homeDir)

	configPath := k8s.ConfigPath(homeDir)
	log.Printf("Derived kubeconfig path: \033[32m%s\033[97m", configPath)

	config := k8s.Config(configPath)
	log.Printf("Successfully found kubeconfig at: \033[32m%s\033[97m", configPath)

	log.Println("Creating kubectl clientset..")
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get Kubernetes client: %s", err)
	}
	return config, client, nil
}

func determineDeployColour(targetEnv, appName string, kubernetes v1.CoreV1Interface) string {
	offlineServiceName := deployment.OfflineServiceName(targetEnv, appName)
	log.Printf("Looking for the colour selector of the offline service: \033[32m%s\033[97m", offlineServiceName)
	service, err := kubernetes.Services("default").Get(offlineServiceName, metav1.GetOptions{})
	if err != nil || service == nil {
		log.Printf("Offline service \033[32m%s\033[97m was not found, defaulting..", offlineServiceName)
	} else if service.Spec.Selector == nil || len(service.Spec.Selector) == 0 {
		if _, ok := service.Spec.Selector["colour"]; !ok {
			log.Printf("Offline service \033[32m%s\033[97m was found but it had no colour selector, defaulting..", offlineServiceName)
		} else {
			log.Printf("Offline service \033[32m%s\033[97m was found but it had no selectors, defaulting..", offlineServiceName)
		}
	}
	return k8s.ServiceSelectorColour(service, err)
}
