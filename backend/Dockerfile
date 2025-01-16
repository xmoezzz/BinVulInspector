FROM golang:1.20 AS builder

ENV CGO_ENABLED=0
ENV GOPROXY=https://goproxy.cn,direct

ARG CI_COMMIT_SHORT_SHA
ARG CI_COMMIT_TAG

WORKDIR /app
COPY . .

RUN go env &&  \
    go mod tidy &&  \
    bash scripts/build.sh native


FROM ubuntu:jammy

COPY --from=builder /app/target/bin-vul-inspector /opt/bin-vul-inspector/bin-vul-inspector
COPY --from=builder /app/config.yaml /opt/bin-vul-inspector/conf/config.yaml

ENV PATH="${PATH}:/opt/bin-vul-inspector"

RUN apt update && \
    apt install -y curl htop && \
    apt clean

WORKDIR /opt/bin-vul-inspector

CMD ["bin-vul-inspector", "server", "--config", "conf/config.yaml"]