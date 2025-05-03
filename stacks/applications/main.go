package main

import (
	"meshed/kubernetes"
	"meshed/stacks/applications/apps"

	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		conf := config.New(ctx, "satellite")

		n2yoKey := conf.RequireSecret("N2YO_KEY")
		longitude := conf.RequireSecret("CURRENT_LONGITUDE")
		latitude := conf.RequireSecret("CURRENT_LATITUDE")

		satelliteConfig := &apps.SatelliteConfig{
			N2YOKey:   n2yoKey,
			Longitude: longitude,
			Latitude:  latitude,
		}

		clusterStack, err := pulumi.NewStackReference(ctx, "scott-the-programmer/meshed/cluster", nil)
		if err != nil {
			return err
		}
		kubeConf := clusterStack.GetStringOutput(pulumi.String("kubeconfig"))

		provider, err := kubernetes.NewKubernetesProvider(ctx, kubeConf)
		if err != nil {
			return err
		}

		appNS, err := corev1.NewNamespace(ctx, "personal",
			&corev1.NamespaceArgs{
				Metadata: metav1.ObjectMetaArgs{
					Name: pulumi.String("personal"),
				},
			}, pulumi.Provider(provider))
		if err != nil {
			return err
		}

		// Create Cloudflared deployment
		cloudflaredArgs := &apps.CloudflaredArgs{
			TunnelSecretName: pulumi.String("cloudflared-token"), // Replace with your secret name
			TunnelSecretKey:  pulumi.String("token"),            // Replace with your secret key
			Subdomain:        pulumi.String("api"),              // Replace with your desired subdomain
		}

		err = apps.NewCloudflared(ctx, provider, appNS, "cloudflared", cloudflaredArgs)
		if err != nil {
			return err
		}

		err = apps.NewBlog(ctx, provider, appNS, "blog")
		if err != nil {
			return err
		}

		err = apps.NewSatellites(ctx, provider, appNS, satelliteConfig, "satellite-api")
		if err != nil {
			return err
		}

		err = apps.NewTermNz(ctx, provider, appNS, "term-nz")
		if err != nil {
			return err
		}

		return nil
	})
}
