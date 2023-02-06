set -e
###
 # @Author: Vincent Young
 # @Date: 2023-02-07 04:55:55
 # @LastEditors: Vincent Young
 # @LastEditTime: 2023-02-07 04:56:15
 # @FilePath: /wb/.cross_compile.sh
 # @Telegram: https://t.me/missuo
 # 
 # Copyright © 2023 by Vincent, All Rights Reserved. 
### 

DIST_PREFIX="weibo"
DEBUG_MODE=${2}
TARGET_DIR="dist"
PLATFORMS="darwin/amd64 darwin/arm64 linux/386 linux/amd64 linux/arm64 linux/mips openbsd/amd64 openbsd/arm64 freebsd/amd64 freebsd/arm64 windows/386 windows/amd64"

rm -rf ${TARGET_DIR}
mkdir ${TARGET_DIR}

for pl in ${PLATFORMS}; do
    export GOOS=$(echo ${pl} | cut -d'/' -f1)
    export GOARCH=$(echo ${pl} | cut -d'/' -f2)
    export TARGET=${TARGET_DIR}/${DIST_PREFIX}_${GOOS}_${GOARCH}
    if [ "${GOOS}" == "windows" ]; then
        export TARGET=${TARGET_DIR}/${DIST_PREFIX}_${GOOS}_${GOARCH}.exe
    fi

    echo "build => ${TARGET}"
    if [ "${DEBUG_MODE}" == "debug" ]; then
        go build -trimpath -gcflags "all=-N -l" -o ${TARGET} \
            -ldflags    "-w -s" main.go
    else
        go build -trimpath -o ${TARGET} \
            -ldflags    "-w -s" main.go
    fi
done