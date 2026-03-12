terraform {
  required_version = ">= 1.0.0"

  required_providers {
    alicloud = {
      source  = "aliyun/alicloud"
      version = "~> 1.214.0"
    }
  }

  backend "oss" {
    bucket = "odcime-terraform-state"
    prefix = "compute/github-profile-readme-postprocessor"
    key = "terraform.tfstate"
    region = "cn-hangzhou"
    encrypt = true
  }
}

locals {
  region = "cn-hangzhou"
  binary_name = "bootstrap"
}

provider "alicloud" {
  region = local.region
}

variable "github_api_token" {
  type = string
  sensitive = true
}

data "archive_file" "function_zip" {
  type        = "zip"
  source_file = "${path.module}/${local.binary_name}"
  output_path = "${path.module}/function_code.zip"
}

resource "alicloud_fc_service" "default" {
  name        = "github-profile-readme-postprocessor"
  publish     = true
}

resource "alicloud_fc_function" "default" {
  service     = alicloud_fc_service.default.name
  name        = "restful"
  runtime     = "custom.debian10"
  handler     = "index.handler"
  memory_size = 512
  timeout     = 120

  filename      = data.archive_file.function_zip.output_path
  code_checksum = data.archive_file.function_zip.output_base64sha256

  environment_variables = {
    APP_GITHUB_TOKEN      = var.github_api_token
    GIN_MODE              = "release"
    APP_LISTEN_PORT       = "9000"
  }
}

resource "alicloud_fc_trigger" "default" {
  service  = alicloud_fc_service.default.name
  function = alicloud_fc_function.default.name
  name     = "apisix-entrypoint"
  type     = "http"

  config = jsonencode({
    authType = "anonymous"
    methods  = ["GET", "POST"]
  })
}

output "fc_invoke_domain" {
  value = "https://${alicloud_fc_service.default.name}.${local.region}.fc.aliyuncs.com/2016-08-15/proxy/${alicloud_fc_function.default.name}/${alicloud_fc_trigger.default.name}/"
}
