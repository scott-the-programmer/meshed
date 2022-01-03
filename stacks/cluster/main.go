package main

import (
	"errors"
	"meshed/linode"
	"meshed/scaleway"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		conf, ok := os.LookupEnv("MESHED_CLOUD")
		if !ok {
			conf = "linode"
		}
		var config pulumi.StringOutput
		var err error
		switch conf {
		case "linode":
			config, err = linode.NewLkeCluster(ctx, "meshed", "1.26")
			if err != nil {
				return err
			}
		case "scaleway":
			config, err = scaleway.NewKapsuleCluster(ctx, "meshed", "1.26.4")
			if err != nil {
				return err
			}
		default:
			return errors.New("invalid cloud provider")
		}

		ctx.Export("kubeconfig", config)
		return nil
	})
}
