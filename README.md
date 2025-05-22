## Requirements

## Requirements

- tested with [Terraform](https://www.terraform.io/downloads.html) 1.10.5.
- [Go](https://golang.org/doc/install) v1.23 (to build the provider plugin)



## Development environment setup 

Prepare `golangci-lint` and `golines` as below:
```
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.64.5
golangci-lint --version

go install github.com/segmentio/golines@latest
go install golang.org/x/vuln/cmd/govulncheck@latest

```



To indicate terraform to use a local binary as a provider, modify `~/.terraformrc` like below:

```
provider_installation {

  dev_overrides {
      "hashicorp.com/edu/eci" = "/home/users/wonjung/iaas/elice-cloud-iaas-terraform-provider/bin"
  }

  direct {}
}
```
Change the path accroding to your environment.


## How to build
```
go mod tidy
make build
```
NOTE: Do not change the name of the compiled binary. The name must follow the following format: `terraform-provider-{NAME}` [ref](https://developer.hashicorp.com/terraform/registry/providers/publishing)




