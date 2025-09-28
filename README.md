# Kubernetes Infrastructure

Hello there! This is my personal project for deploying my Kubernetes infrastructure, which includes my ingress controller and various applications, using Pulumi. I've opted to use Scaleway as the Kubernetes provider and Cloudflare for DNS management - these platforms fit my needs best, providing a balance of functionality, affordability, and ease-of-use.

This project is geared towards my personal use - that includes my blog, my personal projects, and more - but if you're curious and want to try it out yourself or reverse-engineer it, you're more than welcome to! However, please remember that running this project with your own credentials will incur cost. Always ensure you understand the pricing models of any cloud services you're using!

This project uses NGINX Ingress Controller instead of Istio for better compatibility with ARM architecture.

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

The project is configured via Pulumi settings. Set these using the `pulumi config set` command:

- `meshed:CLOUDFLARE_BLOG_ZONE_ID`: Zone ID for the blog on Cloudflare.
- `meshed:CLOUDFLARE_TERM_NZ_ZONE_ID`: Zone ID for the term NZ on Cloudflare.
- `meshed:CLOUDFLARE_LEGACY_ZONE_ID`: Zone ID for the legacy domain on Cloudflare.
- `meshed:CLOUDFLARE_EMAIL`: Your Cloudflare email.
- `meshed:CLOUDFLARE_API_KEY`: Your Cloudflare API key.
- `meshed:MESHED_BLOG_DNS`: The DNS for the blog.
- `meshed:MESHED_TERM_NZ_DNS`: The DNS for term NZ.
- `meshed:MESHED_LEGACY_DNS`: The DNS for the legacy domain.
- `meshed:MESHED_EMAIL`: Email for domain registration.
- `meshed:MESHED_ACME_SECRET`: The ACME secret.
- `meshed:MESHED_STAGING`: A boolean value to indicate if it's staging.

Remember to replace `<value>` with the actual values when setting the configuration.

---

Depending on what MESHED_CLOUD is set to, you will need to configure the [linode](https://www.pulumi.com/registry/packages/linode/installation-configuration/) or [scaleway variables](https://www.pulumi.com/registry/packages/scaleway/installation-configuration/)

## Cloudflared Tunnel Helper

The script `create-cloudflared-tunnel.sh` can create a tunnel for either a subdomain or the root (apex) domain.

Examples:

```
# Subdomain
./create-cloudflared-tunnel.sh blog example.com  # creates blog.example.com

# Root / apex (any of these forms):
./create-cloudflared-tunnel.sh example.com       # creates example.com
./create-cloudflared-tunnel.sh @ example.com     # creates example.com
./create-cloudflared-tunnel.sh root example.com  # creates example.com
./create-cloudflared-tunnel.sh apex example.com  # creates example.com
```

It will:

- Create (or reuse) a Cloudflare tunnel named after the subdomain or the sanitized domain for root.
- Create/update a Kubernetes secret `<name>-cloudflared-file` in the `personal` namespace containing the tunnel credentials.
- Add a DNS route for the hostname if it does not already exist.
- Generate a config file at `~/.cloudflared/config-<id>.yml`.

You can then run the tunnel locally:

```
cloudflared tunnel --config ~/.cloudflared/config-blog.yml run
```

Or for the root domain:

```
cloudflared tunnel --config ~/.cloudflared/config-root.yml run
```
