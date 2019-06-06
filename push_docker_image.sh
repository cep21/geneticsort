#!/bin/bash
set -euo pipefail

STACK_NAME=genetic-sort-stack

function stack_output() {
  # Used from https://stackoverflow.com/questions/41628487/getting-outputs-from-aws-cloudformation-describe-stacks
  local RES=$(aws cloudformation describe-stacks --stack-name ${STACK_NAME} --query "Stacks[0].Outputs[?OutputKey=='${1}'].OutputValue" --output text)
  if [[ -z ${RES} ]]; then
    exit 1
  fi
  echo ${RES}
}

function cli_works() {
	aws sts get-caller-identity &> /dev/null
}

function account_id() {
    aws sts get-caller-identity --query Account --output text
}

function get_aws_region() {
    aws configure get region
}

echo "Verifying cli works ..."
if ! cli_works ; then
	echo "Unable to execute get-caller-identity"
	echo "Are you sure you setup your cli correctly?"
	echo "You may need to set the AWS_PROFILE variable"
	exit 1
fi

REGION=$(get_aws_region)

docker build -t $(stack_output ImageName) .
$(aws ecr get-login --no-include-email --region ${REGION})
docker push $(stack_output ImageName)