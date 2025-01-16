FROM one.reimu.moe/nvidia/cuda:12.4.0-runtime-ubuntu22.04 AS build

RUN sed -i s@/archive.ubuntu.com/@/mirrors.ustc.edu.cn/@g /etc/apt/sources.list && \
    sed -i s@/security.ubuntu.com/@/mirrors.ustc.edu.cn/@g /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends python3 python3-pip python3-venv

RUN pip config set global.index-url https://mirrors.aliyun.com/pypi/simple/
RUN python3 -m venv /bsd-venv

WORKDIR /bsd
COPY torch-1.11.0+cu113-cp310-cp310-linux_x86_64.whl .
RUN /bsd-venv/bin/pip install torch-1.11.0+cu113-cp310-cp310-linux_x86_64.whl
COPY docker-requirements.txt .
RUN /bsd-venv/bin/pip install -r docker-requirements.txt

FROM one.reimu.moe/nvidia/cuda:12.4.0-runtime-ubuntu22.04

RUN sed -i s@/archive.ubuntu.com/@/mirrors.ustc.edu.cn/@g /etc/apt/sources.list && \
    sed -i s@/security.ubuntu.com/@/mirrors.ustc.edu.cn/@g /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends python3 python3-pip python3-venv

COPY --from=build /bsd-venv /bsd-venv
WORKDIR /bsd
COPY . .
RUN rm torch-1.11.0+cu113-cp310-cp310-linux_x86_64.whl && \
    rm docker-requirements.txt
