package main

import (
	"errors"
	"fmt"
	"github.com/Hutchison-Technologies/bluegreen-deployer/cli"
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

var flags = []*cli.Flag{
	&cli.Flag{
		Key:         "chart-dir",
		Default:     "./chart",
		Description: "directory containing the service-to-be-deployed's chart definition.",
		Validator:   IsDirectory,
	},
	&cli.Flag{
		Key:         "app-name",
		Default:     "",
		Description: "name of the service-to-be-deployed (lower-case, alphanumeric + dashes).",
		Validator:   IsValidAppName,
	},
	&cli.Flag{
		Key:         "app-version",
		Default:     "",
		Description: "semantic version of the service-to-be-deployed (vX.X.X, or X.X.X).",
		Validator:   IsValidAppVersion,
	},
	&cli.Flag{
		Key:         "target-env",
		Default:     "",
		Description: "name of the environment in which to deploy the service (prod or staging).",
		Validator:   IsValidTargetEnv,
	},
}

func main() {
	log.Println("Beginning deployment..")

	log.Println("Parsing CLI flags..")
	cliFlags, flagsErr := cli.ParseFlags(flags)
	if flagsErr != nil {
		panic(flagsErr.Error())
	}
	log.Println("Successfully parsed CLI flags:")
	reportMap(cliFlags)

	chartDir, appName, targetEnv, _ := cliFlags["chart-dir"], cliFlags["app-name"], cliFlags["target-env"], cliFlags["app-version"]

	log.Println("Locating chart values..")
	chartValues, chartValuesErr := locateChartValues(chartDir, targetEnv)
	if chartValuesErr != nil {
		panic(chartValuesErr.Error())
	}
	log.Println("Successfully located chart values")

	log.Println("Accessing kubernetes..")
	kubernetes := kubernetesCoreV1()
	log.Println("Determining deploy colour..")
	deployColour := determineDeployColour(targetEnv, appName, kubernetes)
	log.Printf("Determined deploy colour: \033[32m%s\033[97m", deployColour)
	deploymentName := DeploymentName(targetEnv, deployColour, appName)
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
		res, installErr := helmClient.InstallRelease(chartDir, "default", helm.ReleaseName(deploymentName), helm.InstallWait(true), helm.InstallTimeout(300), helm.InstallDescription("Some chart"), helm.ValueOverrides(values))
		if installErr != nil {
			panic(installErr.Error())
		}
		log.Printf("Successfully installed: %s", res.Release.Info.Description)
	} else if releaseContent.Release.Info.Status.Code == release.Status_DELETED {
		log.Println("Existing release found in DELETED state, upgrading chart..")
		res, updateErr := helmClient.UpdateRelease(deploymentName, chartDir, helm.UpgradeForce(true), helm.UpgradeRecreate(true), helm.UpgradeWait(true), helm.UpgradeTimeout(300), helm.UpdateValueOverrides(values))
		if updateErr != nil {
			panic(updateErr.Error())
		}
		log.Printf("Successfully upgraded: %s", res.Release.Info.Description)
	} else {
		log.Println("Existing release found, performing dry-run release..")

		dryRunResponse, dryRunErr := helmClient.UpdateRelease(deploymentName, chartDir, helm.UpgradeDryRun(true), helm.UpgradeForce(true), helm.UpgradeRecreate(true), helm.UpgradeWait(true), helm.UpgradeTimeout(300), helm.UpdateValueOverrides(values))
		if dryRunErr != nil {
			panic(dryRunErr.Error())
		}

		currentManifests := manifest.ParseRelease(releaseContent.Release)
		newManifests := manifest.ParseRelease(dryRunResponse.Release)

		log.Println("Checking proposed release for changes against existing release..")
		hasChanges := diff.DiffManifests(currentManifests, newManifests, []string{}, -1, os.Stderr)
		if hasChanges {
			res, updateErr := helmClient.UpdateRelease(deploymentName, chartDir, helm.UpgradeForce(true), helm.UpgradeRecreate(true), helm.UpgradeWait(true), helm.UpgradeTimeout(300), helm.UpdateValueOverrides(values))
			if updateErr != nil {
				panic(updateErr.Error())
			}
			log.Printf("Successfully upgraded: %s", res.Release.Info.Description)
		} else {
			log.Println("No difference detected, no deploy.")
		}
	}
}

func reportMap(m map[string]string) {
	for key, value := range m {
		log.Printf("\t%s: \033[32m%s\033[97m", key, value)
	}
}

func locateChartValues(chartDir, targetEnv string) (string, error) {
	chartValues := fmt.Sprintf("%s/%s.yaml", chartDir, targetEnv)
	if !FileExists(chartValues) {
		return "", errors.New(fmt.Sprintf("Expected to find chart values yaml at: \033[31m%s\033[97m, but found nothing.", chartValues))
	} else {
		return chartValues, nil
	}
}

func homeDir() string {
	return os.Getenv("HOME")
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
	homeDir := homeDir()
	log.Printf("Using home dir: \033[32m%s\033[97m", homeDir)

	kubeConfigPath := KubeConfigPath(homeDir)
	log.Printf("Derived kubeconfig path: \033[32m%s\033[97m", kubeConfigPath)

	config := KubeConfig(kubeConfigPath)
	log.Printf("Successfully found kubeconfig at: \033[32m%s\033[97m", kubeConfigPath)

	log.Println("Creating kubectl clientset..")
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get Kubernetes client: %s", err)
	}
	return config, client, nil
}

func determineDeployColour(targetEnv, appName string, kubernetes v1.CoreV1Interface) string {
	offlineServiceName := OfflineServiceName(targetEnv, appName)
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
	return ServiceSelectorColour(service, err)
}
