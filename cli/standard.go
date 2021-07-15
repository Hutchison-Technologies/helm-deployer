package cli

import (
	"log"

	"github.com/Hutchison-Technologies/helm-deployer/deployment"
	"github.com/Hutchison-Technologies/helm-deployer/filesystem"
)

func StandardChartFlags() []*Flag {
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
			Key:         TARGET_ENV,
			Default:     "",
			Description: "name of the environment in which to deploy the service (prod or staging).",
			Validator:   deployment.IsValidTargetEnv,
		},
	}
}

func RunStandardChartDeploy() error {
	log.Println("Parsing CLI flags..")
	cliFlags := parseCLIFlags(StandardChartFlags())
	log.Println("Successfully parsed CLI flags:")
	PrintMap(cliFlags)

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
		[][]interface{}{},
		helmConfig,
		cliFlags[CHART_DIR])
	log.Printf("Successfully deployed %s, the service is now live!", Green(deploymentName))
	PrintRelease(deployedRelease)
	return nil
}
