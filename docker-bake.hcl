// Go version
variable "GO_VERSION" {
  default = "1.22"
}

variable "WAYBACK_IPFS_TARGET" {
  default = ""
}

variable "WAYBACK_IPFS_APIKEY" {
  default = ""
}

variable "WAYBACK_IPFS_SECRET" {
  default = ""
}

// GitHub reference as defined in GitHub Actions (eg. refs/head/master))
variable "GITHUB_REF" {
  default = ""
}

target "_common" {
  args = {
    GO_VERSION = GO_VERSION
    WAYBACK_IPFS_TARGET = WAYBACK_IPFS_TARGET
    WAYBACK_IPFS_APIKEY = WAYBACK_IPFS_APIKEY
    WAYBACK_IPFS_SECRET = WAYBACK_IPFS_SECRET
  }
}

group "default" {
  targets = ["image-local"]
}

// Special target: https://github.com/docker/docker-metadata-action#bake-definition
target "docker-metadata-action" {
  tags = ["wabarc/wayback:local"]
}

target "image" {
  inherits = ["_common", "docker-metadata-action"]
}

target "image-local" {
  inherits = ["image"]
  output = ["type=docker"]
}

target "artifact" {
  inherits = ["image"]
  output = ["./dist"]
}

target "artifact-all" {
  inherits = ["artifact"]
  platforms = [
    "linux/386",
    "linux/amd64",
    "linux/arm/v6",
    "linux/arm/v7",
    "linux/arm64",
    "linux/ppc64le"
  ]
}

target "release" {
  inherits = ["image"]
  context = "./"
  platforms = [
    "linux/386",
    "linux/amd64",
    "linux/arm/v6",
    "linux/arm/v7",
    "linux/arm64",
    "linux/ppc64le"
  ]
}

target "bundle" {
  inherits = ["image"]
  context = "./"
  dockerfile = "./build/docker/Dockerfile.all"
  platforms = [
    "linux/386",
    "linux/amd64",
    "linux/arm/v6",
    "linux/arm/v7",
    "linux/arm64",
    "linux/ppc64le"
  ]
}
