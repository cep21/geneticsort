#!/bin/bash
set -exuo pipefail

STACK_NAME=genetic-sort-stack
STACK_FILE=file://cfstack.yaml

# https://stackoverflow.com/questions/949314/how-to-retrieve-the-hash-for-the-current-commit-in-git
export GIT_COMMIT=${GIT_COMMIT-$(git rev-parse --verify HEAD)}

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

function verify() {
	which aws
	verify_cli
	which go
	which golangci-lint
	which docker
}

function account_id() {
    aws sts get-caller-identity --query Account --output text
}

function get_aws_region() {
    aws configure get region
}

function verify_cli() {
    echo "Verifying aws cli works ..."
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

export NUM_JOBS=${NUM_JOBS-2}
export JOB_RUN_TIME=${JOB_RUN_TIME-1m}
export ARRAY_SIZE=${ARRAY_SIZE-1000}

function run_job() {
    aws batch submit-job --job-name geneticsort \
        --job-queue $(stack_output JobQueue) \
        --job-definition $(stack_output JobDefinition) \
        --array-properties "size=${NUM_JOBS}" \
        --container-overrides "environment=[{name=ARRAY_SIZE,value=${ARRAY_SIZE}},{name=RAND_SEED,value=-1},{name=RUN_TIME,value=${JOB_RUN_TIME}}]"
}

function stack_exists() {
	aws cloudformation describe-stacks --stack-name ${STACK_NAME} &> /dev/null
}

function lint() {
    golangci-lint run
}

function fix() {
    find . -iname '*.go' -print0 | xargs -0 gofmt -s -w
    find . -iname '*.go' -print0 | xargs -0 goimports -w
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
        aws cloudformation create-stack \
            --stack-name ${STACK_NAME} \
            --template-body ${STACK_FILE} \
            --capabilities CAPABILITY_NAMED_IAM \
            --parameters ParameterKey=ImageTag,ParameterValue=${GIT_COMMIT}
        echo "Waiting for stack to finish creating"
        aws cloudformation wait stack-create-complete --stack-name ${STACK_NAME}
        echo "Things look good!"
    else
        echo "Stack already exists. Updating and pulling existing information.  If you get a 'No updates are to be performed' error, ignore it"
        aws cloudformation update-stack \
            --stack-name ${STACK_NAME} \
            --template-body ${STACK_FILE} \
            --capabilities CAPABILITY_NAMED_IAM \
            --parameters ParameterKey=ImageTag,ParameterValue=${GIT_COMMIT}
            2> /tmp/err || true
        cat /tmp/err
        if ! grep -q  'No updates' /tmp/err; then
            aws cloudformation wait stack-update-complete --stack-name ${STACK_NAME}
        fi
    fi
}

function go_build() {
    go build ./...
}

function travis() {
    GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint
    go build ./...
    run_test
    lint
}

function is_git_commit_needed {
  [[ $(git diff --shortstat 2> /dev/null | tail -n1) != "" ]]
}


function everything() {
    if is_git_commit_needed; then
        echo "Please run git-commit before you run everything"
        exit 1
    fi
    verify
    go_build
    run_test
    lint
    create_stack
    docker_push
    run_job
}

function present() {
    ~/go/bin/present -base ./present_base/ ./presentation.slide
}

case "${1-}" in
  docker_push)
    docker_push
    ;;
  go_build)
    go_build
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
  everything)
    everything
    ;;
  present)
    present
    ;;
  *)
    echo "Invalid param ${1-}"
    echo "Valid: go_build|docker_push|create_stack|run_job|lint|fix|test|run|travis|everything"
    exit 1
    ;;
esac
