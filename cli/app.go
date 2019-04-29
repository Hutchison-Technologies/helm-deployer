package cli

import (
	"errors"
	"github.com/Hutchison-Technologies/bluegreen-deployer/charts"
	"github.com/Hutchison-Technologies/bluegreen-deployer/deployment"
	"github.com/Hutchison-Technologies/bluegreen-deployer/filesystem"
	"github.com/Hutchison-Technologies/bluegreen-deployer/gosexy/yaml"
	"github.com/Hutchison-Technologies/bluegreen-deployer/k8s"
	"github.com/Hutchison-Technologies/bluegreen-deployer/kubectl"
	"github.com/Hutchison-Technologies/bluegreen-deployer/runtime"
	"github.com/databus23/helm-diff/diff"
	"github.com/databus23/helm-diff/manifest"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/proto/hapi/services"
	"log"
	"os"
)

const (
	CHART_DIR      = "chart-dir"
	APP_NAME       = "app-name"
	APP_VERSION    = "app-version"
	TARGET_ENV     = "target-env"
	DEFAULT_COLOUR = "blue"
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

func Run() error {
	log.Println("Starting bluegreen-deployer..")

	log.Println("Parsing CLI flags..")
	cliFlags := parseCLIFlags(flags)
	log.Println("Successfully parsed CLI flags:")
	PrintMap(cliFlags)

	log.Println("Initialising kubectl..")
	kubeClient := kubeCtlClient()
	log.Println("Successfully initialised kubectl")

	log.Println("Determining deploy colour..")
	deployColour := determineDeployColour(cliFlags[TARGET_ENV], cliFlags[APP_NAME], kubeClient)
	log.Printf("Determined deploy colour: \033[32m%s\033[97m", deployColour)

	//TODO: examine requirements.yaml to ensure blue-green-microservice is a dependency, and it is aliased to bluegreen

	log.Println("Loading chart values..")
	chartValuesYaml := loadChartValues(cliFlags[CHART_DIR], cliFlags[TARGET_ENV])
	log.Println("Successfully loaded chart values")

	log.Println("Setting deployment-specific chart values..")
	chartValues := editChartValues(chartValuesYaml, [][]interface{}{
		[]interface{}{"bluegreen", "is_service_release", false},
		[]interface{}{"bluegreen", "deployment", "colour", deployColour},
		[]interface{}{"bluegreen", "deployment", "version", cliFlags[APP_VERSION]},
	})
	log.Printf("Successfully edited chart values:\n\033[33m%s\033[97m", string(chartValues))

	deploymentName := deployment.DeploymentName(cliFlags[TARGET_ENV], deployColour, cliFlags[APP_NAME])

	log.Println("Connecting helm client..")
	helmClient := buildHelmClient()
	log.Println("Successfully connected helm client!")

	log.Printf("Deploying: \033[32m%s\033[97m..", deploymentName)
	deployedRelease, err := deployRelease(helmClient, deploymentName, cliFlags[CHART_DIR], chartValues)
	runtime.PanicIfError(err)
	log.Printf("Successfully deployed \033[32m%s\033[97m", deploymentName)
	PrintRelease(deployedRelease)

	log.Println("Preparing to route live traffic to newly deployed release")
	log.Println("Setting deployment-specific chart values..")
	chartValues = editChartValues(chartValuesYaml, [][]interface{}{
		[]interface{}{"bluegreen", "is_service_release", true},
		[]interface{}{"bluegreen", "service", "selector", "colour", deployColour},
	})
	log.Printf("Successfully edited chart values:\n\033[33m%s\033[97m", string(chartValues))

	deploymentName = deployment.DeploymentServiceName(cliFlags[TARGET_ENV], cliFlags[APP_NAME])

	log.Printf("Switching service \033[32m%s\033[97m to point at \033[32m%s\033[97m", deploymentName, deployColour)
	deployedRelease, err = deployRelease(helmClient, deploymentName, cliFlags[CHART_DIR], chartValues)
	runtime.PanicIfError(err)
	log.Printf("Successfully switched service \033[32m%s\033[97m to point at \033[32m%s\033[97m", deploymentName, deployColour)
	PrintRelease(deployedRelease)
	return nil
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

func determineDeployColour(targetEnv, appName string, kubeClient v1.CoreV1Interface) string {
	log.Printf("Getting the offline service of \033[32m%s\033[97m in \033[32m%s\033[97m", appName, targetEnv)
	offlineService, err := deployment.GetOfflineService(kubeClient, targetEnv, appName)
	if err != nil {
		log.Println(err.Error())
	}
	if err != nil || offlineService == nil {
		log.Printf("Unable to locate offline service, this might be the first deploy, defaulting to: \033[32m%s\033[97m", DEFAULT_COLOUR)
	} else {
		log.Printf("Found offline service \033[32m%s\033[97m, checking selector colour..", offlineService.GetName())
		offlineColour := k8s.ServiceSelectorColour(offlineService)
		if offlineColour != "" {
			return offlineColour
		}
	}
	return DEFAULT_COLOUR
}

func buildHelmClient() *helm.Client {
	log.Println("Setting up tiller tunnel..")
	tillerHost, err := kubectl.SetupTillerTunnel()
	runtime.PanicIfError(err)
	log.Println("Established tiller tunnel")
	helmClient := helm.NewClient(helm.Host(tillerHost), helm.ConnectTimeout(60))
	log.Printf("Configured helm client, pinging tiller at: \033[32m%s\033[97m..", tillerHost)
	err = helmClient.PingTiller()
	runtime.PanicIfError(err)
	return helmClient
}

func deployRelease(helmClient *helm.Client, releaseName, chartDir string, chartValues []byte) (*release.Release, error) {
	log.Printf("Checking for existing \033[32m%s\033[97m release..", releaseName)
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
		log.Println("Upgrading release..")
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
