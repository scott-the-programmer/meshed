package kubernetes

import (
	"encoding/base64"

	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// NewKubernetesProvider returns an internal kubernetes runtime provider
func NewKubernetesProvider(ctx *pulumi.Context, kc pulumi.StringOutput) (*kubernetes.Provider, error) {
	decoded := kc.ApplyT(func(config string) (string, error) {
		s, err := base64.StdEncoding.DecodeString(config)
		if err != nil {
			return string(config), nil
		}
		return string(s), err
	}).(pulumi.StringOutput)

	p, err := kubernetes.NewProvider(ctx, "internal_provider", &kubernetes.ProviderArgs{Kubeconfig: decoded})
	return p, err
}
