#!/bin/bash

# KubeEye 多架构镜像构建脚本
# 支持 amd64, arm64, arm/v7 架构

set -e

# 配置变量
IMAGE_NAME="${IMAGE_NAME:-kubeeye}"
TAG="${TAG:-latest}"
REGISTRY="${REGISTRY:-}"

# 如果设置了REGISTRY，则添加前缀
if [ -n "$REGISTRY" ]; then
    FULL_IMAGE_NAME="${REGISTRY}/${IMAGE_NAME}"
else
    FULL_IMAGE_NAME="${IMAGE_NAME}"
fi

echo "🏗️  开始构建多架构镜像: ${FULL_IMAGE_NAME}:${TAG}"

# 检查Docker buildx是否可用
if ! docker buildx version > /dev/null 2>&1; then
    echo "❌ Docker buildx 不可用，请确保Docker版本支持buildx"
    exit 1
fi

# 创建buildx构建器实例（如果不存在）
BUILDER_NAME="kubeeye-multiarch"
if ! docker buildx ls | grep -q $BUILDER_NAME; then
    echo "📦 创建多架构构建器..."
    docker buildx create --name $BUILDER_NAME --use
else
    echo "📦 使用已存在的构建器: $BUILDER_NAME"
    docker buildx use $BUILDER_NAME
fi

# 启用binfmt_misc支持（用于交叉编译）
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes

echo "🔨 构建支持的架构: linux/amd64, linux/arm64"

# 构建并推送多架构镜像
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --tag "${FULL_IMAGE_NAME}:${TAG}" \
    --tag "${FULL_IMAGE_NAME}:latest" \
    --push \
    .

echo "✅ 多架构镜像构建完成!"
echo "📦 镜像标签: ${FULL_IMAGE_NAME}:${TAG}"
echo "📦 镜像标签: ${FULL_IMAGE_NAME}:latest"

# 显示镜像信息
echo ""
echo "🔍 镜像详细信息:"
docker buildx imagetools inspect "${FULL_IMAGE_NAME}:${TAG}"
