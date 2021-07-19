package cli

import (
	"errors"
	"fmt"
	"log"

	"github.com/Hutchison-Technologies/helm-deployer/charts"
	"github.com/Hutchison-Technologies/helm-deployer/deployment"
	"github.com/Hutchison-Technologies/helm-deployer/filesystem"
	"github.com/Hutchison-Technologies/helm-deployer/k8s"
	"github.com/Hutchison-Technologies/helm-deployer/runtime"
)

func BlueGreenFlags() []*Flag {
	return []*Flag{
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
}

func RunBlueGreenDeploy() error {
	log.Println("Parsing CLI flags..")
	cliFlags := parseCLIFlags(BlueGreenFlags())
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

	log.Println("Configuring helm...")
	helmConfig := buildHelmConfig()
	log.Println("Successfully configured helm!")

	deploymentName := deployment.BlueGreenDeploymentName(cliFlags[TARGET_ENV], deployColour, cliFlags[APP_NAME]) //
	log.Printf("Preparing to deploy %s..", Green(deploymentName))
	deployedRelease := releaseWithValues(
		deploymentName,
		chartValuesYaml,
		deployment.ChartValuesForDeployment(deployColour, cliFlags[APP_VERSION]),
		helmConfig,
		cliFlags[CHART_DIR])

	log.Println("Now updating the online deployment replica set to a minimum of 1.")
	scaleOnlineReplicaSetResult := scaleReplicaSet(deploymentName, 1)
	if scaleOnlineReplicaSetResult != nil {
		panic(fmt.Errorf("Failed to scale replica set HPA: %v", scaleOnlineReplicaSetResult))
	}
	log.Printf("Successfully deployed %s", Green(deploymentName))
	PrintRelease(deployedRelease)

	log.Println("For the deployment to go live, the service selector colour will be updated")
	serviceDeploymentName := deployment.ServiceReleaseName(cliFlags[TARGET_ENV], cliFlags[APP_NAME])
	log.Printf("Preparing to deploy %s..", Green(serviceDeploymentName))
	deployedServiceRelease := releaseWithValues(
		serviceDeploymentName,
		chartValuesYaml,
		deployment.ChartValuesForServiceRelease(deployColour),
		helmConfig,
		cliFlags[CHART_DIR])
	log.Printf("Successfully deployed %s, the service is now live!", Green(serviceDeploymentName))
	PrintRelease(deployedServiceRelease)

	log.Println("To reduce costing, number of pods in offline deployments will now be scaled to zero.")
	currentOfflineColour := determineDeployColour(cliFlags[TARGET_ENV], cliFlags[APP_NAME])
	log.Printf("Offline colour is %s", currentOfflineColour)

	// Build strings for offline deployment and autoscaler
	offlineDeploymentName := deployment.BlueGreenDeploymentName(cliFlags[TARGET_ENV], currentOfflineColour, cliFlags[APP_NAME])
	offlineHPAName := fmt.Sprintf("%s-hpa", offlineDeploymentName)

	log.Printf("We will first remove the Horizontal Pod Autoscaler (%s) from the offline service.", offlineHPAName)
	deletionResult := deleteHPA(offlineHPAName)
	if deletionResult != nil {
		log.Printf("Failed to delete HPA: %v", deletionResult)
		log.Println("This can happen if this is a  first deployment; skipping.")
	}

	log.Println("Now updating the offline service replica set to zero.")
	scaleReplicaSetResult := scaleReplicaSet(offlineDeploymentName, 0)
	if scaleReplicaSetResult != nil {
		log.Printf("Failed to scale replica set HPA: %v", scaleReplicaSetResult)
		log.Println("This can happen if this is a  first deployment; skipping.")
	}
	log.Println("Updates complete!")

	return nil
}

func assertChartIsBlueGreen(chartDir string) {
	chartYamlPath := charts.ChartYamlPath(chartDir)
	log.Printf("Checking %s for blue-green-microservice dependency..", Green(chartYamlPath))
	hasBlueGreenDependency := charts.HasDependency(chartYamlPath, "blue-green-microservice", "bluegreen")
	if !hasBlueGreenDependency {
		runtime.PanicIfError(errors.New(fmt.Sprintf("Dependency %s must be present and aliased to %s in the %s file in order to deploy using this program.", Green("blue-green-microservice"), Green("bluegreen"), Green(chartYamlPath))))
	}
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
