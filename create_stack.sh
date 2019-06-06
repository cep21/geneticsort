#!/bin/bash
set -euo pipefail

STACK_NAME=genetic-sort-stack
STACK_FILE=file://cfstack.yaml

function cli_works() {
    # A safe command that should work on any correctly configured CLI
	aws sts get-caller-identity &> /dev/null
}

# Returns non zero if the stack already exists
function stack_exists() {
	aws cloudformation describe-stacks --stack-name ${STACK_NAME} &> /dev/null
}

if ! cli_works ; then
    echo "CLI may not be configured correctly"
	exit 1
fi

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
