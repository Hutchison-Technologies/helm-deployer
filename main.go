package main

import (
	"flag"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"log"
	"os"
	"time"
)

func main() {
	log.Println("Beginning deployment")

	chartDir, chartValues, appName, targetEnv, appVersion := parseCmdLineFlags()
	log.Printf(
		`Preparing to deploy using these variables:
			chartDir: %s
			chartValues: %s
			appName: %s
			targetEnv: %s
			appVersion: %s`,
		chartDir, chartValues, appName, targetEnv, appVersion)

	kubernetes := kubernetesCoreV1()

	offlineServiceName := OfflineServiceName(targetEnv, appName)
	log.Printf("Looking for the offline service: %s", offlineServiceName)
	for {
		service, err := kubernetes.Services("default").Get(appName, metav1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}
		// fmt.Printf("There are %d service in the cluster\n", len(service.Items))
		log.Printf("%v", service)
		// log.Printf("%v", service.Selector)
		// Examples for error handling:
		// - Use helper functions like e.g. errors.IsNotFound()
		// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
		namespace := "default"
		pod := "example-xxxxx"
		_, err = kubernetes.Pods(namespace).Get(pod, metav1.GetOptions{})
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
	invalidFlags := false
	if *chartDir == "" || !IsDirectory(*chartDir) {
		invalidFlags = true
		log.Printf("Invalid chart-dir: %s, must be %s", *chartDir, chartDirUsage)
	}
	if *appName == "" || !IsValidAppName(*appName) {
		invalidFlags = true
		log.Printf("Invalid app-name: %s, must be %s", *appName, appNameUsage)
	}
	if *appVersion == "" || !IsValidAppVersion(*appVersion) {
		invalidFlags = true
		log.Printf("Invalid app-version: %s, must be %s", *appVersion, appVersionUsage)
	}
	if *targetEnv == "" || !IsValidTargetEnv(*targetEnv) {
		invalidFlags = true
		log.Printf("Invalid target-env: %s, must be %s", *targetEnv, targetEnvUsage)
	}
	if chartValues == "" || !FileExists(chartValues) {
		invalidFlags = true
		log.Printf("Expected to find chart values yaml at: %s, but found nothing.", chartValues)
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
	log.Printf("Using home dir: %s", homeDir)

	kubeConfigPath := KubeConfigPath(homeDir)
	log.Printf("Derived kubeconfig path: %s", kubeConfigPath)

	config := KubeConfig(kubeConfigPath)
	log.Printf("Successfully found kubeconfig at: %s", kubeConfigPath)

	log.Println("Creating kubectl clientset..")
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	log.Println("Successfully created kubectl clientset")
	return clientset.CoreV1()
}
