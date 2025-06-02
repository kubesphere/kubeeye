FROM python:3.12-slim

LABEL maintainer="KubeSphere Team"
LABEL description="KubeEye Kubernetes Cluster Inspection Tool"

# 设置工作目录
WORKDIR /app

# 安装系统依赖
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    openssh-client \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# 复制项目文件
COPY . /app/

# 安装Python依赖
RUN pip install --no-cache-dir -r requirements.txt

# 创建数据目录
RUN mkdir -p /app/data/clusters /app/data/results /app/data/logs

# 设置环境变量
ENV PYTHONPATH=/app
ENV KUBEEYE_DATA_DIR=/app/data

# 暴露服务端口
EXPOSE 8501

# 初始化应用
RUN python init.py

# 启动应用
ENTRYPOINT ["streamlit", "run", "app.py", "--server.port=8501", "--server.address=0.0.0.0"]
