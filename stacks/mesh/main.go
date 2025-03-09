package main

import (
	"meshed/kubernetes"
	"meshed/kubernetes/resources"
	"meshed/local"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		var kubeConfig pulumi.StringOutput
		var err error

		localCluster := os.Getenv("MESHED_LOCAL") == "true"
		if localCluster {
			kubeConfig, err = local.NewLocalCluster(ctx)
			if err != nil {
				return err
			}
		} else {
			clusterStack, err := pulumi.NewStackReference(ctx, "scott-the-programmer/meshed/cluster", nil)
			if err != nil {
				return err
			}
			kubeConfig = clusterStack.GetStringOutput(pulumi.String("kubeconfig"))
		}
		conf := config.New(ctx, "")

		blogDnsZoneId := conf.RequireSecret("CLOUDFLARE_BLOG_ZONE_ID")
		termNzDnsZoneId := conf.RequireSecret("CLOUDFLARE_TERM_NZ_ZONE_ID")
		legacyDnsZoneId := conf.RequireSecret("CLOUDFLARE_LEGACY_ZONE_ID")
		cloudflareEmail := conf.RequireSecret("CLOUDFLARE_EMAIL")
		blogDns := conf.Get("MESHED_BLOG_DNS")
		termNzDns := conf.Get("MESHED_TERM_NZ_DNS")
		apiDns := conf.Get("MESHED_API_DNS")
		legacyDns := conf.Get("MESHED_LEGACY_DNS")
		email := conf.RequireSecret("MESHED_EMAIL")
		acmeSecret := conf.RequireSecret("MESHED_ACME_SECRET")
		//staging := conf.GetBool("MESHED_STAGING")

		replacer := resources.NewReplacer("../../")

		pulumi.All(blogDnsZoneId, termNzDnsZoneId, legacyDnsZoneId, cloudflareEmail, email, acmeSecret).ApplyT(func(values []interface{}) error {
			replacer.Add("MESHED_BLOG_DNS", blogDns)
			replacer.Add("MESHED_TERM_NZ_DNS", termNzDns)
			replacer.Add("MESHED_API_DNS", apiDns)
			replacer.Add("MESHED_LEGACY_DNS", legacyDns)
			replacer.Add("CLOUDFLARE_BLOG_ZONE_ID", values[0].(string))
			replacer.Add("CLOUDFLARE_TERM_NZ_ZONE_ID", values[1].(string))
			replacer.Add("CLOUDFLARE_LEGACY_ZONE_ID", values[2].(string))
			replacer.Add("CLOUDFLARE_EMAIL", values[3].(string))
			replacer.Add("MESHED_EMAIL", values[4].(string))
			replacer.Add("MESHED_ACME_SECRET", values[5].(string))

			provider, err := kubernetes.NewKubernetesProvider(ctx, kubeConfig)
			if err != nil {
				return err
			}

			// Setup basic ingress instead of Istio mesh
			err = kubernetes.NewIngressController(ctx, provider, replacer)
			if err != nil {
				return err
			}

			// certMan, err := kubernetes.NewCertManager(ctx, gw, provider, replacer)
			// if err != nil {
			// 	return err
			// }

			// err = kubernetes.NewCerts(ctx, provider, replacer, certMan, staging)
			// if err != nil {
			// 	return err
			// }

			// blogDnsZoneId.ApplyT(func(val string) error {
			// 	_, err = cloudflare.NewARecord(ctx, cloudflare.ARecordArgs{Source: blogDns, Target: ip.Elem().ToStringOutput(), ZoneID: val, Proxied: false})
			// 	if err != nil {
			// 		return err
			// 	}

			// 	_, err = cloudflare.NewARecord(ctx, cloudflare.ARecordArgs{Source: apiDns, Target: ip.Elem().ToStringOutput(), ZoneID: val, Proxied: false})
			// 	if err != nil {
			// 		return err
			// 	}

			// 	_, err = cloudflare.NewARecord(ctx, cloudflare.ARecordArgs{Source: fmt.Sprintf("%s.%s", "www", blogDns), Target: ip.Elem().ToStringOutput(), ZoneID: val, Proxied: false})
			// 	if err != nil {
			// 		return err
			// 	}
			// 	return nil
			// })

			// termNzDnsZoneId.ApplyT(func(val string) error {
			// 	_, err = cloudflare.NewARecord(ctx, cloudflare.ARecordArgs{Source: termNzDns, Target: ip.Elem().ToStringOutput(), ZoneID: val, Proxied: false})
			// 	if err != nil {
			// 		return err
			// 	}
			// 	_, err = cloudflare.NewARecord(ctx, cloudflare.ARecordArgs{Source: fmt.Sprintf("%s.%s", "www", termNzDns), Target: ip.Elem().ToStringOutput(), ZoneID: val, Proxied: false})
			// 	if err != nil {
			// 		return err
			// 	}
			// 	return nil
			// })

			// legacyDnsZoneId.ApplyT(func(val string) error {
			// 	_, err = cloudflare.NewARecord(ctx, cloudflare.ARecordArgs{Source: legacyDns, Target: ip.Elem().ToStringOutput(), ZoneID: val, Proxied: false})
			// 	if err != nil {
			// 		return err
			// 	}
			// 	_, err = cloudflare.NewARecord(ctx, cloudflare.ARecordArgs{Source: fmt.Sprintf("%s.%s", "www", legacyDns), Target: ip.Elem().ToStringOutput(), ZoneID: val, Proxied: false})
			// 	if err != nil {
			// 		return err
			// 	}
			// 	return nil
			// })
			return nil
		})

		return nil
	})
}
