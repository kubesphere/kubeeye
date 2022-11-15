#!/bin/bash

set -e

GV="bolingcavalry:v1"

rm -rf ./pkg/clients
./hack/generate_group.sh "client,lister,informer" k8s_customize_controller/pkg/clients k8s_customize_controller/pkg/apis "${GV}" --output-base=./  -h "$PWD/hack/boilerplate.go.txt" -v 10
mv k8s_customize_controller/pkg/clients ./pkg/
rm -rf ./k8s_customize_controller
