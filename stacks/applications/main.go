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

		provider, err := kubernetes.NewLocalKubernetesProvider(ctx)
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
		cloudflaredBlogArgs := &apps.CloudflaredArgs{
			TunnelName:       pulumi.String("blog-tunnel"),
			Subdomain:        pulumi.String("scott"),
			Domain:           pulumi.String("murray.kiwi"),
		}

		cloudflaredBlogApiArgs := &apps.CloudflaredArgs{
			TunnelName:       pulumi.String("posts-api-tunnel"),
			Subdomain:        pulumi.String("blog-api"),
			Domain:           pulumi.String("murray.kiwi"),
		}

		cloudflaredSatelliteArgs := &apps.CloudflaredArgs{
			TunnelName:       pulumi.String("blog-api-tunnel"),
			Subdomain:        pulumi.String("api"),
			Domain:           pulumi.String("murray.kiwi"),
		}

		cloudflaredTermNzArgs := &apps.CloudflaredArgs{
			TunnelName:       pulumi.String("term-nz-tunnel"),
			Domain:           pulumi.String("term.nz"),
		}

		blogArgs := &apps.BlogArgs{
			Cloudflared: cloudflaredBlogArgs,
		}

		err = apps.NewBlog(ctx, provider, appNS, "blog", blogArgs)
		if err != nil {
			return err
		}

		blogApiArgs := &apps.BlogApiArgs{
			Cloudflared: cloudflaredBlogApiArgs,
		}

		err = apps.NewBlogApi(ctx, provider, appNS, "blog-api", blogApiArgs)
		if err != nil {
			return err
		}

		satellitesArgs := &apps.SatellitesArgs{
			SatelliteConfig: satelliteConfig,
			Cloudflared:     cloudflaredSatelliteArgs,
		}

		err = apps.NewSatellites(ctx, provider, appNS, satellitesArgs, "satellite-api")
		if err != nil {
			return err
		}

		termNzArgs := &apps.TermNzArgs{
			Cloudflared: cloudflaredTermNzArgs,
		}

		err = apps.NewTermNz(ctx, provider, appNS, "term-nz", termNzArgs)
		if err != nil {
			return err
		}

		return nil
	})
}
