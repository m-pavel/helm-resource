package cmd

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	cv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func GetQuota(namespace string) (*cv1.ResourceQuota, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	pc := clientset.CoreV1().ResourceQuotas(namespace)
	rql, err := pc.List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	if len(rql.Items) == 0 {
		return nil, fmt.Errorf("no resource quotas defined in namespace %s", namespace)
	}
	if len(rql.Items) > 1 {
		return nil, fmt.Errorf("%d resource quotas defined in namespace %s", len(rql.Items), namespace)
	}
	return &rql.Items[0], nil
}
