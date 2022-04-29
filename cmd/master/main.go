package main

import (
	"context"
	"github.com/zdarovich/cowboy_shooters/internal/cowboy"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	cowboyPath = flag.String("cowboys", "cowboys.json", "Path to cowboys list")
)

// Master deploys log pod and cowboys pods. Cowboy pod information is read from json file
func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	// build configuration from the config file.
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	// create kubernetes clientset. this clientset can be used to create,delete,patch,list etc for the kubernetes resources
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	cowboys, err := GetCowboysFromJson(*cowboyPath)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()

	logPod := GetLogPod()
	logPod, err = clientset.CoreV1().Pods(logPod.Namespace).Create(ctx, logPod, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	logPodIp := ""
	for len(logPodIp) == 0 {
		fmt.Println("Log pod ip lookup...")
		logPod, err = clientset.CoreV1().Pods(logPod.Namespace).Get(ctx, logPod.Name, metav1.GetOptions{})
		if err != nil {
			panic(err)
		}
		logPodIp = logPod.Status.PodIP
	}
	fmt.Println("Pod created successfully...")

	logServiceAddr := fmt.Sprintf("%s:%d", logPodIp, logPod.Spec.Containers[0].Ports[0].ContainerPort)
	for i, cowboy := range cowboys {
		// build the pod definition we want to deploy
		pod := GetCowboyPod(cowboy.Name, cowboy.Health, cowboy.Damage, logServiceAddr, 50002+i)
		// now create the pod in kubernetes cluster using the clientset
		pod, err = clientset.CoreV1().Pods(pod.Namespace).Create(ctx, pod, metav1.CreateOptions{})
		if err != nil {
			panic(err)
		}
		fmt.Println("Pod created successfully...")
	}

}

func GetLogPod() *core.Pod {
	return &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "log-pod",
			Namespace: "default",
			Labels: map[string]string{
				"app": "log",
			},
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:            "log",
					Image:           "log:latest",
					ImagePullPolicy: core.PullIfNotPresent,
					Ports: []core.ContainerPort{
						{
							Name:          "http",
							HostPort:      50001,
							ContainerPort: 50001,
							Protocol:      core.ProtocolTCP,
						},
					},
					Env: []core.EnvVar{
						{
							Name:  "PORT",
							Value: "50001",
						},
					},
				},
			},
		},
	}
}

func GetCowboyPod(name string, health, damage int, logServiceAddr string, port int) *core.Pod {
	return &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.ToLower(name),
			Namespace: "default",
			Labels: map[string]string{
				"app":  "cowboy",
				"name": name,
			},
		},
		Spec: core.PodSpec{
			RestartPolicy: core.RestartPolicyNever,
			Containers: []core.Container{
				{
					Name:            strings.ToLower(name),
					Image:           "cowboy:latest",
					ImagePullPolicy: core.PullIfNotPresent,
					Ports: []core.ContainerPort{
						{
							Name:          "http",
							HostPort:      int32(port),
							ContainerPort: int32(port),
							Protocol:      core.ProtocolTCP,
						},
					},
					Env: []core.EnvVar{
						{
							Name:  "LOG_ADDR",
							Value: logServiceAddr,
						},
						{
							Name:  "NAME",
							Value: name,
						},
						{
							Name:  "HEALTH",
							Value: strconv.Itoa(health),
						},
						{
							Name:  "DAMAGE",
							Value: strconv.Itoa(damage),
						},
						{
							Name:  "PORT",
							Value: strconv.Itoa(port),
						},
					},
				},
			},
		},
	}
}

func GetCowboysFromJson(path string) ([]*cowboy.Cowboy, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	var cowboys []*cowboy.Cowboy
	if err := json.NewDecoder(fd).Decode(&cowboys); err != nil {
		return nil, errors.Wrap(err, "Failed to decode file: ")
	}

	return cowboys, nil
}
