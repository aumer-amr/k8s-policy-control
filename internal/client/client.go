package client

import (
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	c client.Client
)

func InitClient(config *rest.Config) error {
	cl, err := client.New(config, client.Options{})
	if err != nil {
		return err
	}

	c = cl
	return nil
}

func Client() client.Client {
	return c
}
