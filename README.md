# ecrgate

## WIP
- Builds docker image (done)
- Creates ECR repo if it does not exist (done)
- Uploads docker image to ECR (done)
- Pulls scan results of image (done)
- Compares scan results to a yaml file with acceptable thresholds
- Returns a non zero exit code if results are above thresholds
- Optionally deletes the docker tag from the repo if the scan is above threshold

## Requirements
- docker daemon running
- aws credentials (with access to ECR)
