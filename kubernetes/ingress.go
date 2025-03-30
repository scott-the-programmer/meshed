package kubernetes

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// GetIngressIP returns the IP address of the NGINX ingress controller
func GetIngressIP(ctx *pulumi.Context, provider *kubernetes.Provider, ingressRelease *helm.Release) pulumi.StringPtrOutput {
	service, err := v1.GetService(ctx, "ingress-nginx/ingress-nginx-controller", ingressRelease.ID(), nil, pulumi.Provider(provider), pulumi.Parent(ingressRelease))
	if err != nil {
		return pulumi.StringPtrOutput{}
	}

	loadBalancerIp := service.Status.ApplyT(func(status *v1.ServiceStatus) *string {
		if status == nil || status.LoadBalancer == nil || len(status.LoadBalancer.Ingress) == 0 {
			return nil
		}
		return status.LoadBalancer.Ingress[0].Ip
	}).(pulumi.StringPtrOutput)

	return loadBalancerIp
}
