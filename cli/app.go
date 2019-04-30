package cli

import (
	"errors"
	"fmt"
	"github.com/Hutchison-Technologies/bluegreen-deployer/charts"
	"github.com/Hutchison-Technologies/bluegreen-deployer/deployment"
	"github.com/Hutchison-Technologies/bluegreen-deployer/filesystem"
	"github.com/Hutchison-Technologies/bluegreen-deployer/gosexy/yaml"
	"github.com/Hutchison-Technologies/bluegreen-deployer/h3lm"
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
	CHART_DIR               = "chart-dir"
	APP_NAME                = "app-name"
	APP_VERSION             = "app-version"
	TARGET_ENV              = "target-env"
	DEFAULT_COLOUR          = "blue"
	RELEASE_UPGRADE_TIMEOUT = 300
	ROLLBACK_VERSION_POOL   = 50
	ROLLBACK_TIMEOUT        = 600
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

	log.Println("Asserting that this is a bluegreen microservice chart..")
	assertChartIsBlueGreen(cliFlags[CHART_DIR])
	log.Println("This is a bluegreen microservice chart!")

	log.Println("Determining deploy colour..")
	deployColour := determineDeployColour(cliFlags[TARGET_ENV], cliFlags[APP_NAME])
	log.Printf("Determined deploy colour: %s", Green(deployColour))

	log.Println("Loading chart values..")
	chartValuesYaml := loadChartValues(cliFlags[CHART_DIR], cliFlags[TARGET_ENV])
	log.Println("Successfully loaded chart values")

	log.Println("Connecting helm client..")
	helmClient := buildHelmClient()
	log.Println("Successfully connected helm client!")

	deploymentName := deployment.DeploymentName(cliFlags[TARGET_ENV], deployColour, cliFlags[APP_NAME])
	log.Printf("Preparing to deploy %s..", Green(deploymentName))
	deployedRelease := releaseWithValues(
		deploymentName,
		chartValuesYaml,
		deployment.ChartValuesForDeployment(deployColour, cliFlags[APP_VERSION]),
		helmClient,
		cliFlags[CHART_DIR])
	log.Printf("Successfully deployed %s", Green(deploymentName))
	PrintRelease(deployedRelease)

	log.Println("For the deployment to go live, the service selector colour will be updated")
	serviceDeploymentName := deployment.ServiceReleaseName(cliFlags[TARGET_ENV], cliFlags[APP_NAME])
	log.Printf("Preparing to deploy %s..", Green(serviceDeploymentName))
	deployedServiceRelease := releaseWithValues(
		serviceDeploymentName,
		chartValuesYaml,
		deployment.ChartValuesForServiceRelease(deployColour),
		helmClient,
		cliFlags[CHART_DIR])
	log.Printf("Successfully deployed %s, the service is now live!", Green(serviceDeploymentName))
	PrintRelease(deployedServiceRelease)
	return nil
}

func parseCLIFlags(flagsToParse []*Flag) map[string]string {
	potentialParsedFlags, potentialParseFlagsErr := ParseFlags(flagsToParse)
	cliFlags, err := HandleParseFlags(potentialParsedFlags, potentialParseFlagsErr)
	runtime.PanicIfError(err)
	return cliFlags
}

func assertChartIsBlueGreen(chartDir string) {
	requirementsYamlPath := charts.RequirementsYamlPath(chartDir)
	log.Printf("Checking %s for blue-green-microservice dependency..", Green(requirementsYamlPath))
	hasBlueGreenDependency := charts.HasDependency(requirementsYamlPath, "blue-green-microservice", "bluegreen")
	if !hasBlueGreenDependency {
		runtime.PanicIfError(errors.New(fmt.Sprintf("Dependency %s must be present and aliased to %s in the %s file in order to deploy using this program.", Green("blue-green-microservice"), Green("bluegreen"), Green(requirementsYamlPath))))
	}
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

func determineDeployColour(targetEnv, appName string) string {
	log.Println("Initialising kubectl..")
	kubeClient := kubeCtlClient()
	log.Println("Successfully initialised kubectl")

	log.Printf("Getting the offline service of %s in %s", Green(appName), Green(targetEnv))
	offlineService, err := deployment.GetOfflineService(kubeClient, targetEnv, appName)
	if err != nil {
		log.Println(err.Error())
	}
	if err != nil || offlineService == nil {
		log.Printf("Unable to locate offline service, this might be the first deploy, defaulting to: %s", Green(DEFAULT_COLOUR))
	} else {
		log.Printf("Found offline service %s, checking selector colour..", Green(offlineService.GetName()))
		offlineColour := k8s.ServiceSelectorColour(offlineService)
		if offlineColour != "" {
			return offlineColour
		}
	}
	return DEFAULT_COLOUR
}

func kubeCtlClient() v1.CoreV1Interface {
	client, err := kubectl.Client()
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
