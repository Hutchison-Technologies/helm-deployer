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
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
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

	log.Println("Connecting helm client..")
	helmClient := buildHelmClient()
	log.Println("Successfully connected helm client!")

	deploymentName := deployment.BlueGreenDeploymentName(cliFlags[TARGET_ENV], deployColour, cliFlags[APP_NAME])
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

	log.Println("To reduce costing, number of pods in offline deployments will now be scaled to zero.")
	currentOfflineColour := determineDeployColour(cliFlags[TARGET_ENV], cliFlags[APP_NAME])
	deploymentsClient := kubeCtlAppClient().Deployments(apiv1.NamespaceDefault)
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Build string to search for
		offlineDeploymentName := deployment.BlueGreenDeploymentName(cliFlags[TARGET_ENV], currentOfflineColour, cliFlags[APP_NAME])

		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver

		result, getErr := deploymentsClient.Get(offlineDeploymentName, metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("Failed to get latest version of Deployment: %v", getErr))
		}

		var numberOfReplicas int32 = 0
		result.Spec.Replicas = &numberOfReplicas
		_, updateErr := deploymentsClient.Update(result)
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("Update failed: %v", retryErr))
	}
	log.Println("Updated deployment scaling to zero replicas...")
	log.Println("Updates complete!")

	return nil
}

func assertChartIsBlueGreen(chartDir string) {
	requirementsYamlPath := charts.RequirementsYamlPath(chartDir)
	log.Printf("Checking %s for blue-green-microservice dependency..", Green(requirementsYamlPath))
	hasBlueGreenDependency := charts.HasDependency(requirementsYamlPath, "blue-green-microservice", "bluegreen")
	if !hasBlueGreenDependency {
		runtime.PanicIfError(errors.New(fmt.Sprintf("Dependency %s must be present and aliased to %s in the %s file in order to deploy using this program.", Green("blue-green-microservice"), Green("bluegreen"), Green(requirementsYamlPath))))
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
