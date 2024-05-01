# STEP1: CLONE THE CODE
FROM --platform=$BUILDPLATFORM alpine:latest as builder_prepare
WORKDIR /home/
RUN apk add git && \
    git clone --depth 1 https://github.com/mickael-kerjean/filestash

# STEP2: BUILD THE FRONTEND
FROM --platform=$BUILDPLATFORM node:18-alpine AS builder_frontend
WORKDIR /home/
COPY --from=builder_prepare /home/filestash/ ./
RUN apk add make git gzip brotli && \
    npm install --legacy-peer-deps && \
    make build_frontend && \
    cd public && make compress

# STEP3: BUILD THE BACKEND
# 使用 buildx 的 TARGETPLATFORM 环境变量来选择正确的基础镜像
FROM --platform=$BUILDPLATFORM golang:1.20 AS builder_backend
WORKDIR /home/
COPY --from=builder_frontend /home/ ./
# 使用条件语句根据不同的平台执行不同的命令
RUN if [ "$(uname -m)" = "aarch64" ]; then \
        apt-get update && apt-get install -y libvips-dev curl make libjpeg-dev libtiff-dev libpng-dev libwebp-dev libraw-dev libheif-dev libgif-dev; \
    else \
        apt-get update && apt-get install -y libvips-dev curl make libjpeg-dev libtiff-dev libpng-dev libwebp-dev libraw-dev libheif-dev libgif-dev; \
    fi && \
    make build_init && \
    make build_backend && \
    mkdir -p ./dist/data/state/config/ && \
    cp config/config.json ./dist/data/state/config/config.json

# STEP4: Create the prod image from the build
FROM --platform=$BUILDPLATFORM debian:stable-slim
COPY --from=builder_backend /home/dist/ /app/
WORKDIR "/app"
RUN apt-get update && apt-get install -y --no-install-recommends apt-utils curl emacs-nox ffmpeg zip poppler-utils && \
    # org-mode: html export
    curl https://raw.githubusercontent.com/mickael-kerjean/filestash/master/server/.assets/emacs/htmlize.el > /usr/share/emacs/site-lisp/htmlize.el && \
    # org-mode: markdown export
    curl https://raw.githubusercontent.com/mickael-kerjean/filestash/master/server/.assets/emacs/ox-gfm.el > /usr/share/emacs/site-lisp/ox-gfm.el && \
    # Cleanup
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    rm -rf /tmp/*

RUN useradd filestash && \
    chown -R filestash:filestash /app/ && \
    chmod 730 /app/filestash

USER filestash
CMD ["/app/filestash"]
EXPOSE 8334
