#!/bin/sh

# This command is used in the consuemr to download the static videos from s3

CONTAINER_ALREADY_STARTED="CONTAINER_ALREADY_STARTED_PLACEHOLDER"
if [ ! -e $CONTAINER_ALREADY_STARTED ]; then
    touch $CONTAINER_ALREADY_STARTED
    echo "-- First container startup downloading s3 files --"
    aws s3 cp s3://lecture-processor/background/ ./static --recursive
else
    echo "-- Not first container startup resume s3 files--"
fi

./main