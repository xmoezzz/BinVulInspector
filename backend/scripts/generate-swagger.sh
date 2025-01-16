#!/usr/bin/env bash
set -eux

# v1.16.3
swag -v
swag fmt --exclude scs-workdir

go version
go generate bin-vul-inspector/cmd/server