package client

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

type ListOptions struct {
	ApiVersion string
	Kind       string
	Name       string
	Namespace  string
	Selector   labels.Selector
}

type KubeClient struct {
	Client client.Client
}

func New(config *rest.Config) error {
	k := &KubeClient{}
	cl, err := client.New(config, client.Options{})
	if err != nil {
		return err
	}

	k.Client = cl

	return nil
}
