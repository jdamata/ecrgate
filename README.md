[![CircleCI](https://circleci.com/gh/jdamata/ecrgate.svg?style=svg)](https://circleci.com/gh/jdamata/ecrgate)
[![codecov](https://codecov.io/gh/jdamata/ecrgate/branch/master/graph/badge.svg)](https://codecov.io/gh/jdamata/ecrgate)
[![Go Report Card](https://goreportcard.com/badge/github.com/jdamata/ecrgate)](https://goreportcard.com/report/github.com/jdamata/ecrgate)
[![GPLv3 License](https://img.shields.io/badge/License-GPL%20v3-yellow.svg)](https://opensource.org/licenses/)

# ecrgate
ecrgate is used to simplify the building, pushing and scanning of docker images into AWS ECR. It can build docker iamges, create AWS ECR repositories, push docker images, check AWS ECR scan results, etc...

The main usage for this tool is in CI pipelines where we want to fail a pipeline if a docker image does not pass specific thresholds of vulnerabilities.

## Installation

Linux:
```bash
wget https://github.com/jdamata/ecrgate/releases/latest/download/ecrgate-linux-amd64
chmod u+x ecrgate-linux-amd64
mv ecrgate-linux-amd64 ~/bin/ecrgate
```

Mac:
```bash
wget https://github.com/jdamata/ecrgate/releases/latest/download/ecrgate-darwin-amd64
chmod u+x ecrgate-darwin-amd64
mv ecrgate-darwin-amd64 ~/bin/ecrgate
```

## Flags
--repo is the only required flag.

```bash
$ go run main.go --help
Build, push and gate docker image promotion to ECR

Usage:
  ecrgate [flags]

Flags:
  -a, --accounts strings    List of AWS account ids to allow pulling images from
  -c, --clean               Delete image from ECR if scan fails threshold
      --critical int        Acceptable threshold for CRITICAL level results
  -d, --dockerfile string   Path to Dockerfile (default ".")
  -h, --help                help for ecrgate
  -i, --image               Existing docker image to pull down instead of building a new one
      --high int            Acceptable threshold for HIGH level results (default 3)
      --info int            Acceptable threshold for INFORMATIONAL level results (default 25)
      --low int             Acceptable threshold for LOW level results (default 10)
      --medium int          Acceptable threshold for MEDIUM level results (default 5)
  -r, --repo string         ECR repo to create and push image to
  -t, --tag string          Docker tag to build (default "latest")
  -s, --disable_scan        Skip checking AWS ECR scan results
      --version             version for ecrgate
```

## Examples

```bash
# Use ecrgate defaults and local dir for Dockerfile
ecrgate --repo joel-test

# Specify path to Dockerfile, docker tag and delete image on failed scan
ecrgate --repo joel-test --dockerfile example/ --tag $(git describe --abbrev=0 --tags) --clean

# Specify threshold levels
ecrgate --repo joel-test --dockerfile example/ --tag $(git rev-parse --short HEAD) --clean \ 
  --info 10 --low 5 --medium 3 --high 2 --critical 1

# Use a remote image instead of building a local one
ecrgate --repo ingress-nginx --image quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.30.0
```

## Requirements
- Docker
- AWS credentials

Sample IAM policy:
```
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Stmt1585513157885",
      "Action": [
        "ecr:BatchDeleteImage",
        "ecr:CreateRepository",
        "ecr:DescribeImageScanFindings",
        "ecr:DescribeRepositories",
        "ecr:GetAuthorizationToken",
        "ecr:PutImageScanningConfiguration",
        "ecr:PutImageTagMutability"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
```