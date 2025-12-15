FROM python:3.9-slim

WORKDIR /app

# 替换源以加速构建 (尝试替换旧版 sources.list 和新版 debian.sources)
RUN (sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list || true) \
    && (sed -i 's/security.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list || true) \
    && (sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list.d/debian.sources || true)

# 安装依赖 (使用清华源加速)
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple

# 字体支持
RUN apt-get update && apt-get install -y fonts-wqy-zenhei && rm -rf /var/lib/apt/lists/*

# 复制代码
COPY main.py .
COPY plotter.py .

# 启动
CMD ["python", "-u", "main.py"]
