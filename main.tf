terraform {
  required_providers {
    alicloud = {
      source  = "aliyun/alicloud"
      version = "~> 1.214.0"
    }
  }

  backend "oss" {
    bucket = "odcime-terraform-state"
    prefix = "compute/github-profile-readme-postprocessor"
    key    = "terraform.tfstate"
    region = "cn-hangzhou"
  }
}

provider "alicloud" {}

variable "github_api_token" {
  type      = string
  sensitive = true
}

data "alicloud_account" "current" {}

data "archive_file" "bootstrap_zip" {
  type        = "zip"
  source_file = "${path.module}/bootstrap"
  output_path = "${path.module}/bootstrap.zip"
}

resource "alicloud_fc_service" "default" {
  name    = "github-profile-readme-postprocessor"
  publish = true
}

resource "alicloud_fc_function" "default" {
  service     = alicloud_fc_service.default.name
  name        = "restful"
  runtime     = "custom.debian10"
  memory_size = 128
  timeout     = 15

  filename         = data.archive_file.bootstrap_zip.output_path
  source_code_hash = data.archive_file.bootstrap_zip.output_base64sha256

  custom_runtime_config {
    command = ["./bootstrap"]
  }

  environment_variables = {
    FC_CUSTOM_LISTEN_PORT = "8080"  # Tell FC runtime to listen on 8080.
    APP_GITHUB_TOKEN          = var.github_api_token
    GIN_MODE              = "release"
  }
}

resource "alicloud_fc_trigger" "default" {
  service       = alicloud_fc_service.default.name
  function      = alicloud_fc_function.default.name
  name          = "apisix-entrypoint"
  type          = "http"
  config_mocker = false
  config = <<EOF
  {
      "authType": "anonymous",
      "methods": ["GET"]
  }
  EOF
}

output "fc_invoke_domain" {
  value = "${data.alicloud_account.current.id}.cn-hangzhou.fc.aliyuncs.com"
}
