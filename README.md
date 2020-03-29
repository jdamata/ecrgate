[![GPLv3 License](https://img.shields.io/badge/License-GPL%20v3-yellow.svg)](https://opensource.org/licenses/)
[![GitHub Release](https://img.shields.io/github/release/jdamata/ecrgate.svg?style=flat)](https://github.com/jdamata/ecrgate/releases?sort-semver)
[![Go Report Card](https://goreportcard.com/badge/github.com/jdamata/ecrgate)](https://goreportcard.com/report/github.com/jdamata/ecrgate)
[![Downloads](https://img.shields.io/github/downloads/jdamata/ecrgate/total.svg)](https://github.com/jdamata/ecrgate/releases)

# ecrgate
ecrgate is used to add a security gate to your CI pipeline.  
This flow of the utility is as follows:  
- Create the specified ECR repo (If it does not exist already)
- Build, tag and push your Dockerfile to the ECR repo. 
- Pull down the scan results of that image
- Compare them to the thresholds specified in flags (or defaults)
- Return exit code 1 if thresholds are too high
- (Optional) delete the docker image from the ECR repo if 

## Running


## Requirements
- Docker
- AWS credentials

## Flags
--repo is the only required flag.

```bash
$ go run main.go --help
Build, push and gate docker image promotion to ECR

Usage:
  ecrgate [flags]

Flags:
  -c, --clean               Delete image from ECR if scan fails threshold
      --critical int        Acceptable threshold for CRITICAL level results
  -d, --dockerfile string   Path to Dockerfile (default ".")
  -h, --help                help for ecrgate
      --high int            Acceptable threshold for HIGH level results (default 3)
      --info int            Acceptable threshold for INFORMATIONAL level results (default 25)
      --low int             Acceptable threshold for LOW level results (default 10)
      --medium int          Acceptable threshold for MEDIUM level results (default 5)
  -r, --repo string         ECR repo to create and push image to
  -t, --tag string          Docker tag to build (default "latest")
      --version             version for ecrgate
```