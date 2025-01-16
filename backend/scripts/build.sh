#!/usr/bin/env sh

set -e

BUILD_TIME=${BUILDTIME:-$(TZ=Asia/Shanghai date --rfc-3339=sec)}

GO_LDFLAGS=${GO_LDFLAGS:-"-w -s"}

# 定义要添加的链接变量，带上 -X
declare -a flags
flags=(
  "-X 'bin-vul-inspector/cmd/version.GitCommit=${CI_COMMIT_SHORT_SHA:-unknown}'"
  "-X 'bin-vul-inspector/cmd/version.BuildTime=${BUILD_TIME}'"
  "-X 'bin-vul-inspector/cmd/version.Version=${CI_COMMIT_TAG:-unknown}'"
)

# 添加每个变量到 GO_LDFLAGS
GO_LDFLAGS="$GO_LDFLAGS ${flags[@]}"

build() {
  CGO_ENABLED=0 GOOS=$1 GOARCH=$2 go build --trimpath --ldflags "$GO_LDFLAGS" -o "$3" "$4"
}

build_native() {
  mkdir -p target/

  CGO_ENABLED=0 go build --trimpath --ldflags "$GO_LDFLAGS" -o target/bin-vul-inspector bin-vul-inspector/cmd/server
}

build_linux_amd64() {
  mkdir -p target/linux/amd64

  build linux amd64 target/linux/amd64/bin-vul-inspector bin-vul-inspector/cmd/server
}

build_linux_arm64() {
  mkdir -p target/linux/arm64

  build linux arm64 target/linux/arm64/bin-vul-inspector bin-vul-inspector/cmd/server
}

case $1 in
"native")
  build_native
;;
"cross-all")
  build_linux_amd64
  build_linux_arm64
  ;;
"linux-amd64")
  build_linux_amd64
  ;;
"linux-arm64")
  build_linux_arm64
  ;;
"prune")
  rm -rf target
  ;;
*)
  echo "Usage: $0 native/linux-amd64/linux-arm64/cross-all"
  ;;
esac
