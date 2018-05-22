FROM alpine

# 创建应用目录
RUN mkdir -p /usr/src/app

# 安装依赖
WORKDIR /usr/src/app
COPY qrserve .

CMD ["./qrserve", "0.0.0.0:8080"]
