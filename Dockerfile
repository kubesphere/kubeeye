FROM python:3.12-alpine

LABEL maintainer="KubeSphere Team"
LABEL description="KubeEye Kubernetes Cluster Inspection Tool"
LABEL version="2.0.0"

# 设置构建参数以支持多架构
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

# 设置工作目录
WORKDIR /app

# 安装系统依赖
RUN apk add --no-cache \
    gcc \
    musl-dev \
    libffi-dev \
    openssl-dev \
    openssh-client \
    curl \
    git \
    && apk upgrade --no-cache

# 首先复制requirements.txt以利用Docker缓存
COPY requirements.txt /app/

# 安装Python依赖
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir -r requirements.txt

# 复制项目文件
COPY . /app/

# 创建数据目录并设置权限
RUN mkdir -p /app/data/clusters /app/data/results /app/data/logs /app/data/schedules /app/data/git_rules

# 设置环境变量
ENV PYTHONPATH=/app
ENV KUBEEYE_DATA_DIR=/app/data
ENV STREAMLIT_SERVER_HEADLESS=true
ENV STREAMLIT_SERVER_ENABLE_CORS=false
ENV STREAMLIT_SERVER_ENABLE_XSRF_PROTECTION=false


# 根据目标架构下载对应的OPA二进制文件
RUN set -eux; \
    OPA_VERSION="v1.5.1"; \
    case "${TARGETARCH}" in \
        amd64) \
            OPA_ARCH="amd64"; \
            ;; \
        arm64) \
            OPA_ARCH="arm64"; \
            ;; \
        arm) \
            OPA_ARCH="arm"; \
            ;; \
        *) \
            echo "Unsupported architecture: ${TARGETARCH}"; \
            exit 1; \
            ;; \
    esac; \
    curl -sSL "https://github.com/open-policy-agent/opa/releases/download/${OPA_VERSION}/opa_linux_${OPA_ARCH}_static" -o opa && \
    chmod +x opa && \
    mv opa /usr/local/bin/opa


# 初始化应用（在切换用户前执行）
RUN python init.py

# 暴露服务端口
EXPOSE 8501

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8501/_stcore/health || exit 1

# 启动应用
ENTRYPOINT ["streamlit", "run", "app.py", "--server.port=8501", "--server.address=0.0.0.0"]
