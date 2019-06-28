/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"github.com/jasonlvhit/gocron"
	"github.com/yangyongzhi/sym-operator/pkg/helm"
	"k8s.io/client-go/rest"
	"os"
	"time"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	clientset "github.com/yangyongzhi/sym-operator/pkg/client/clientset/versioned"
	informers "github.com/yangyongzhi/sym-operator/pkg/client/informers/externalversions"
	"github.com/yangyongzhi/sym-operator/pkg/signals"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	// Enable logs
	klog.InitFlags(flag.NewFlagSet(os.Args[0], flag.ExitOnError))

	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()
	var cfg *rest.Config
	var err error
	if kubeconfig != "" {
		cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kubeconfig) // out of cluster config
	} else {
		cfg, err = rest.InClusterConfig()
	}

	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	helmClient, err := helm.NewClient(cfg, kubeClient)
	if err != nil {
		klog.Fatalf("Error building helm client: %s", err.Error())
	}

	symClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building symphony clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	symInformerFactory := informers.NewSharedInformerFactory(symClient, time.Second*30)

	controller := NewController(kubeClient, symClient, helmClient,
		kubeInformerFactory.Apps().V1().Deployments(),
		//symInformerFactory.Example().V1().Foos()
		symInformerFactory.Devops().V1().Migrates())
	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopCh)
	symInformerFactory.Start(stopCh)

	go func() { <-gocron.Start() }()

	if err = controller.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
