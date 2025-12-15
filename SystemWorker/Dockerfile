FROM python:3.9-slim

WORKDIR /app

# 安装依赖
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# 字体支持 (为了 matplotlib 显示中文)
RUN apt-get update && apt-get install -y fonts-wqy-zenhei && rm -rf /var/lib/apt/lists/*

# 复制代码
COPY main.py .

# 启动
CMD ["python", "-u", "main.py"]
