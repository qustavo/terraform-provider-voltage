terraform {
  required_providers {
    voltage = {
      source = "registry.terraform.io/qustavo/voltage"
    }
  }
}

provider "voltage" {}

resource "voltage_node" "testnet" {
  network = "testnet"
  purchased_type = "ondemand"
  type = "lite"
  name = "qustavo"
  settings = {
    autopilot = false
    grpc = true
    rest = true
    keysend = true
    whitelist = [""]
    alias = "qustavo"
    color = "#000000"
  }
}
