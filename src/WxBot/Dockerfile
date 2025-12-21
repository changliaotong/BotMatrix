# 使用官方 Python 3.9 轻量级镜像 (基于 Debian 11 Bullseye)
FROM python:3.9-slim-bullseye

# 设置工作目录
WORKDIR /app

# 设置环境变量
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

# 优化：更换为阿里云镜像源，加速国内下载
RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list && \
    sed -i 's/security.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list

# 安装系统依赖
# unixodbc-dev: 用于 pyodbc 编译
# freetds-dev: 用于 pymssql 编译 (作为 pyodbc 的 fallback)
# build-essential: 用于编译某些 Python 包
# curl/gnupg: 用于安装 docker-cli
RUN apt-get update && apt-get install -y \
    build-essential \
    unixodbc-dev \
    freetds-dev \
    curl \
    gnupg \
    && rm -rf /var/lib/apt/lists/*

# 复制依赖文件并安装
COPY requirements.txt .
# 优化：使用阿里云 PyPI 镜像加速 pip 安装
RUN pip install --no-cache-dir -r requirements.txt -i https://mirrors.aliyun.com/pypi/simple/

# 复制项目代码
COPY . .

# 暴露端口 (WebUI: 5000, Bot: 3001)
EXPOSE 5000 3001

# 启动命令
CMD ["python", "onebot.py"]
