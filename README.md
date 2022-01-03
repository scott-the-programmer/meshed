# Meshed Infrastructure

Hello there! This is my personal project for deploying my Kubernetes infrastructure, which includes my mesh network and various applications, using Pulumi. I've opted to use Scaleway as the Kubernetes provider and Cloudflare for DNS management - these platforms fit my needs best, providing a balance of functionality, affordability, and ease-of-use.

This project is geared towards my personal use - that includes my blog, my personal projects, and more - but if you're curious and want to try it out yourself or reverse-engineer it, you're more than welcome to! However, please remember that running this project with your own credentials will incur cost. Always ensure you understand the pricing models of any cloud services you're using!

## Getting Started

First things first, you'll want to ensure you've got all the necessary tools installed:

- Pulumi CLI
- Go
- Docker
- kubectl
- helm

Once that's all ready, you can use the various commands I've written into the Makefile to build, test, and deploy your infrastructure. Here's a quick rundown:

- `make build` - Tidies up the Go dependencies and installs any that are missing.
- `make cluster-preview` - Gives you a preview of what changes will be made to the cluster.
- `make cluster-up` - Brings the cluster online.
- `make cluster-down` - Takes the cluster offline.
- `make mesh-preview` - Gives you a preview of what changes will be made to the mesh.
- `make mesh-up` - Brings the mesh online.
- `make mesh-down` - Takes the mesh offline.
- `make apps-preview` - Gives you a preview of what changes will be made to the apps.
- `make apps-up` - Brings the apps online.
- `make apps-down` - Takes the apps offline.
- `make load-config` - Loads the kubeconfig from the cluster stack.

## Configuration

There are also a number of environment variables you'll need to set up. These variables are key to making sure everything is working just right:

| Environment Variable | Description                                                                                  |
| -------------------- | -------------------------------------------------------------------------------------------- |
| `CLOUDFLARE_ZONE_ID` | This is the Zone ID for my domain in Cloudflare. It's used when creating DNS records.        |
| `CLOUDFLARE_EMAIL`   | My Cloudflare account email. This is what's used to authenticate with the Cloudflare API.    |
| `CLOUDFLARE_API_KEY` | My Cloudflare API key. This is also used to authenticate with the Cloudflare API.            |
| `MESHED_BLOG_DNS`    | The domain I use for my blog. It's used when creating the DNS record and Kubernetes Ingress. |
| `MESHED_EMAIL`       | My email that I use with Let's Encrypt for certificate generation.                           |
| `MESHED_ACME_SECRET` | The secret used for the ACME challenge.                                                      |
| `MESHED_CLOUD`       | Currently supports "linode" or "scaleway" (defaults to linode)                               |

Depending on what MESHED_CLOUD is set to, you will need to configure the [linode](https://www.pulumi.com/registry/packages/linode/installation-configuration/) or [scaleway variables](https://www.pulumi.com/registry/packages/scaleway/installation-configuration/)

Feel free to give it a whirl, and see what you can learn from it! And, of course, if you have any suggestions or improvements, I'd love to hear them.
