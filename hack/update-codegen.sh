#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
SCRIPT_ROOT="${SCRIPT_DIR}/.."
CODEGEN_PKG="${CODEGEN_PKG:-"${SCRIPT_ROOT}/vendor/k8s.io/code-generator"}"

echo "Verifying environment..."
echo "SCRIPT_ROOT: ${SCRIPT_ROOT}"
echo "CODEGEN_PKG: ${CODEGEN_PKG}"

# 验证必要文件存在
if [ ! -f "${CODEGEN_PKG}/kube_codegen.sh" ]; then
    echo "Error: kube_codegen.sh not found at ${CODEGEN_PKG}/kube_codegen.sh"
    exit 1
fi

if [ ! -f "${SCRIPT_ROOT}/hack/boilerplate.go.txt" ]; then
    echo "Creating empty boilerplate.go.txt"
    touch "${SCRIPT_ROOT}/hack/boilerplate.go.txt"
fi

if [ ! -d "${SCRIPT_ROOT}/apis/kubeeye" ]; then
    echo "Error: APIs directory not found at ${SCRIPT_ROOT}/apis/kubeeye"
    exit 1
fi

source "${CODEGEN_PKG}/kube_codegen.sh"

THIS_PKG="github.com/kubesphere/kubeeye"

echo "Generating deepcopy functions..."
kube::codegen::gen_helpers \
    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
    "${SCRIPT_ROOT}"

echo "Generating client code..."
kube::codegen::gen_client \
    --with-watch \
    --with-applyconfig \
    --output-dir "${SCRIPT_ROOT}/clients" \
    --output-pkg "${THIS_PKG}/clients" \
    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
    "${SCRIPT_ROOT}/apis"

