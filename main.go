package main

import (
	"flag"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/portforwarder"
	"log"
	"os"
	"time"
)

func main() {
	log.Println("Beginning deployment")

	chartDir, chartValues, appName, targetEnv, appVersion := parseCmdLineFlags()
	log.Println("Preparing to deploy using these variables:")
	log.Printf("\tchartDir: \033[32m%s\033[97m", chartDir)
	log.Printf("\tchartValues: \033[32m%s\033[97m", chartValues)
	log.Printf("\tappName: \033[32m%s\033[97m", appName)
	log.Printf("\ttargetEnv: \033[32m%s\033[97m", targetEnv)
	log.Printf("\tappVersion: \033[32m%s\033[97m", appVersion)

	log.Println("Accessing kubernetes..")
	kubernetes := kubernetesCoreV1()
	log.Println("Determining deploy colour..")
	deployColour := determineDeployColour(targetEnv, appName, kubernetes)
	log.Printf("Determined deploy colour: \033[32m%s\033[97m", deployColour)
	deploymentName := DeploymentName(targetEnv, deployColour, appName)
	log.Printf("Deploying: \033[32m%s\033[97m", deploymentName)
	helmClient()

	for {
		// fmt.Printf("There are %d service in the cluster\n", len(service.Items))
		// Examples for error handling:
		// - Use helper functions like e.g. errors.IsNotFound()
		// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
		namespace := "default"
		pod := "example-xxxxx"
		_, err := kubernetes.Pods(namespace).Get(pod, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("Error getting pod %s in namespace %s: %v\n",
				pod, namespace, statusError.ErrStatus.Message)
		} else if err != nil {
			panic(err.Error())
		} else {
			fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
		}

		time.Sleep(10 * time.Second)
	}
}

func parseCmdLineFlags() (string, string, string, string, string) {
	chartDirUsage := "directory containing the service-to-be-deployed's chart definition."
	chartDir := flag.String("chart-dir", "./chart", chartDirUsage)
	appNameUsage := "name of the service-to-be-deployed (lower-case, alphanumeric + dashes)."
	appName := flag.String("app-name", "", appNameUsage)
	appVersionUsage := "semantic version of the service-to-be-deployed (vX.X.X, or X.X.X)."
	appVersion := flag.String("app-version", "", appVersionUsage)
	targetEnvUsage := "name of the environment in which to deploy the service (prod or staging)."
	targetEnv := flag.String("target-env", "", targetEnvUsage)
	flag.Parse()
	chartValues := fmt.Sprintf("%s/%s.yaml", *chartDir, *targetEnv)
	invalidFlags := *chartDir == "" || !IsDirectory(*chartDir) || *appName == "" || !IsValidAppName(*appName) || *appVersion == "" || !IsValidAppVersion(*appVersion) || *targetEnv == "" || !IsValidTargetEnv(*targetEnv) || chartValues == "" || !FileExists(chartValues)
	if *chartDir == "" {
		log.Printf("Missing flag \033[32m-chart-dir\033[97m, must be \033[33m%s\033[97m", chartDirUsage)
	} else if !IsDirectory(*chartDir) {
		log.Printf("Invalid \033[32mchart-dir\033[97m: \033[31m%s\033[97m, must be \033[33m%s\033[97m", *chartDir, chartDirUsage)
	}
	if *appName == "" {
		log.Printf("Missing flag \033[32m-app-name\033[97m, must be \033[33m%s\033[97m", appNameUsage)
	} else if !IsValidAppName(*appName) {
		log.Printf("Invalid \033[32mapp-name\033[97m: \033[31m%s\033[97m, must be \033[33m%s\033[97m", *appName, appNameUsage)
	}
	if *appVersion == "" {
		log.Printf("Missing flag \033[32m-app-version\033[97m, must be \033[33m%s\033[97m", appVersionUsage)
	} else if !IsValidAppVersion(*appVersion) {
		log.Printf("Invalid \033[32mapp-version\033[97m: \033[31m%s\033[97m, must be \033[33m%s\033[97m", *appVersion, appVersionUsage)
	}
	if *targetEnv == "" {
		log.Printf("Missing flag \033[32m-target-env\033[97m, must be \033[33m%s\033[97m", targetEnvUsage)
	} else if !IsValidTargetEnv(*targetEnv) {
		log.Printf("Invalid \033[32mtarget-env\033[97m: \033[31m%s\033[97m, must be \033[33m%s\033[97m", *targetEnv, targetEnvUsage)
	}
	if !FileExists(chartValues) {
		log.Printf("Expected to find chart values yaml at: \033[31m%s\033[97m, but found nothing.", chartValues)
	}
	if invalidFlags {
		panic("Invalid flag supplied, see log.")
	}
	return *chartDir, chartValues, *appName, *targetEnv, *appVersion
}

func homeDir() string {
	return os.Getenv("HOME")
}

func kubernetesCoreV1() v1.CoreV1Interface {
	homeDir := homeDir()
	log.Printf("Using home dir: \033[32m%s\033[97m", homeDir)

	kubeConfigPath := KubeConfigPath(homeDir)
	log.Printf("Derived kubeconfig path: \033[32m%s\033[97m", kubeConfigPath)

	config := KubeConfig(kubeConfigPath)
	log.Printf("Successfully found kubeconfig at: \033[32m%s\033[97m", kubeConfigPath)

	log.Println("Creating kubectl clientset..")
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	log.Println("Successfully created kubectl clientset")
	return clientset.CoreV1()
}

func helmClient() *helm.Client {
	homeDir := homeDir()
	log.Printf("Using home dir: \033[32m%s\033[97m", homeDir)

	kubeConfigPath := KubeConfigPath(homeDir)
	log.Printf("Derived kubeconfig path: \033[32m%s\033[97m", kubeConfigPath)

	// config := KubeConfig(kubeConfigPath)
	log.Printf("Successfully found kubeconfig at: \033[32m%s\033[97m", kubeConfigPath)
	helmHost, err := setupConnection()
	log.Printf("Using host \033[32m%s\033[97m from config to connect with Helm..", helmHost)

	helmClient := helm.NewClient(helm.Host(helmHost), helm.ConnectTimeout(60))
	log.Println("made client")
	releases, err := helmClient.ListReleases()
	if err != nil {
		panic(err.Error())
	}
	log.Printf("%v", releases)
	return helmClient
}

func setupConnection() (string, error) {
	config, client, err := getKubeClient()
	if err != nil {
		return "", err
	}

	tillerTunnel, err := portforwarder.New("kube-system", client, config)
	if err != nil {
		return "", err
	}

	// tillerHost := ""
	tillerHost := fmt.Sprintf("127.0.0.1:%d", tillerTunnel.Local)
	// log.Printf("Created tunnel using local port: '%d'\n", tillerTunnel.Local)

	// Set up the gRPC config.
	log.Printf("SERVER: %q\n", tillerHost)

	// Plugin support.
	return tillerHost, nil
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
