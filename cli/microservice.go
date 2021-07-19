package cli

import (
	"errors"
	"fmt"
	"log"

	"github.com/Hutchison-Technologies/helm-deployer/charts"
	"github.com/Hutchison-Technologies/helm-deployer/deployment"
	"github.com/Hutchison-Technologies/helm-deployer/filesystem"
	"github.com/Hutchison-Technologies/helm-deployer/runtime"
)

func MicroserviceFlags() []*Flag {
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

func RunMicroserviceDeploy() error {
	log.Println("Parsing CLI flags..")
	cliFlags := parseCLIFlags(MicroserviceFlags())
	log.Println("Successfully parsed CLI flags:")
	PrintMap(cliFlags)

	log.Println("Asserting that this is a microservice chart..")
	assertChartIsMicroservice(cliFlags[CHART_DIR])
	log.Println("This is a microservice chart!")

	log.Println("Loading chart values..")
	chartValuesYaml := loadChartValues(cliFlags[CHART_DIR], cliFlags[TARGET_ENV])
	log.Println("Successfully loaded chart values")

	log.Println("Connecting helm config..")
	helmConfig := buildHelmConfig()
	log.Println("Successfully configured helm!")

	deploymentName := deployment.StandardChartDeploymentName(cliFlags[TARGET_ENV], cliFlags[APP_NAME])
	log.Printf("Preparing to deploy %s..", Green(deploymentName))
	deployedRelease := releaseWithValues(
		deploymentName,
		chartValuesYaml,
		deployment.ChartValuesForMicroserviceDeployment(cliFlags[APP_VERSION]),
		helmConfig,
		cliFlags[CHART_DIR])
	log.Printf("Successfully deployed %s, the service is now live!", Green(deploymentName))
	PrintRelease(deployedRelease)

	return nil
}

func assertChartIsMicroservice(chartDir string) {
	chartYamlPath := charts.ChartYamlPath(chartDir)
	log.Printf("Checking %s for microservice dependency..", Green(chartYamlPath))
	hasDependency := charts.HasDependency(chartYamlPath, "microservice", "")
	if !hasDependency {
		runtime.PanicIfError(errors.New(fmt.Sprintf("Dependency %s must be present in the %s file in order to deploy using this program.", Green("microservice"), Green(chartYamlPath))))
	}
}
