#!/bin/bash
readelf -Ws /usr/lib/x86_64-linux-gnu/libc.so.6 | grep 'FUNC' |  grep -v 'GLIBC_PRIVATE' |awk '{print $8}' | cut -d '@' -f1 > glibc_symbols.txt