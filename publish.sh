#!/bin/sh
SERVICE_NAME="jsdistributed"
BIN_NAME="jsdistributed"

# Build and deploy using the variable
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o "$BIN_NAME" && \
    ssh ops@10.6.0.1 "supervisorctl stop $SERVICE_NAME" && \
    scp "$BIN_NAME" ops@10.6.0.1:/opt/service/"$BIN_NAME"/ && \
    rsync -avz static/ ops@10.6.0.1:/opt/service/"$BIN_NAME"/static && \
    scp config.yaml ops@10.6.0.1:/opt/service/"$BIN_NAME"/ 

ssh ops@10.6.0.1 "supervisorctl start $SERVICE_NAME"   