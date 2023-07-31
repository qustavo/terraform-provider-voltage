# terraform-provider-voltage
A Voltage Cloud Terraform provider

**This project is still under development and it won't be published to terraform registry yet**

## Instructions
Build and install the Go binary:
```bash
make
```

This should create a binary called `terraform-provider-voltage` inside your `go env GOBIN` directory.

Update your `~/.terraformrc` to override the provider installation so that terraform can find the binary you've just installed.
```terraform
provider_installation {
  dev_overrides {
    "registry.terraform.io/qustavo/voltage" = "<YOUR GOBIN PATH>"
  }

  direct {}
}
```

Finally, let's create a node.
```bash
export VOLTAGE_TOKEN=<your token> # or wait for terraform to ask for it.
cd examples

terraform plan # Creates the plan
terraform apply # Creates a node
terraform destroy # Destroys the node
```
