#!/bin/bash
set -exuo pipefail

STACK_NAME=genetic-sort-stack
STACK_FILE=file://cfstack.yaml

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

function verify_cli() {
    echo "Verifying cli works ..."
    if ! cli_works ; then
        echo "Unable to execute get-caller-identity"
        echo "Are you sure you setup your cli correctly?"
        echo "You may need to set the AWS_PROFILE variable"
        exit 1
    fi
}

function docker_push() {
    REGION=$(get_aws_region)
    docker build -t $(stack_output ImageName) .
    $(aws ecr get-login --no-include-email --region ${REGION})
    docker push $(stack_output ImageName)
}

function run_job() {
    aws batch submit-job --job-name geneticsort --job-queue $(stack_output JobQueue) --job-definition $(stack_output JobDefinition)
}

function stack_exists() {
	aws cloudformation describe-stacks --stack-name ${STACK_NAME} &> /dev/null
}

function lint() {
    golangci-lint run
}

function fix() {
    find . -iname '*.go'-print0 | xargs -0 gofmt -s -w
    find . -iname '*.go'-print0 | xargs -0 goimports -w
}

function run_test() {
    env "GORACE=halt_on_error=1" go test -count=1 -v -race ./...
}

function run_go() {
    go run main.go
}

function create_stack() {
    verify_cli
    if ! stack_exists ; then
        echo "stack does not already exist.  Creating"
        aws cloudformation create-stack --stack-name ${STACK_NAME} --template-body ${STACK_FILE} --capabilities CAPABILITY_NAMED_IAM
        echo "Waiting for stack to finish creating"
        aws cloudformation wait stack-create-complete --stack-name ${STACK_NAME}
        echo "Things look good!"
    else
        echo "Stack already exists. Updating and pulling existing information.  If you get a 'No updates are to be performed' error, ignore it"
        aws cloudformation update-stack --stack-name ${STACK_NAME} --template-body ${STACK_FILE} --capabilities CAPABILITY_NAMED_IAM 2> /tmp/err || true
        cat /tmp/err
        if ! grep -q  'No updates' /tmp/err; then
            aws cloudformation wait stack-update-complete --stack-name ${STACK_NAME}
        fi
    fi
}

function travis() {
    GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint
    go build ./...
    run_test
    lint
}

case "${1-}" in
  docker_push)
    docker_push
    ;;
  create_stack)
    create_stack
    ;;
  run_job)
    run_job
    ;;
  lint)
    lint
    ;;
  fix)
    fix
    ;;
  test)
    run_test
    ;;
  run)
    run_go
    ;;
  travis)
    travis
    ;;
  *)
    echo "Invalid param ${1-}"
    echo "Valid: docker_push|create_stack|run_job|lint|fix|test|run|travis"
    exit 1
    ;;
esac
