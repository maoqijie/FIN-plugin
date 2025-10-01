#!/bin/bash

# 跨平台插件构建脚本
# 支持构建多个平台的插件可执行文件

PLUGIN_NAME="example-plugin"

# 定义平台列表
PLATFORMS=(
    "windows/amd64"
    "windows/arm64"
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "android/arm64"
)

echo "开始构建跨平台插件: $PLUGIN_NAME"

# 创建 dist 目录
mkdir -p dist

# 遍历平台进行构建
for PLATFORM in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$PLATFORM"

    OUTPUT_NAME="$PLUGIN_NAME"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi

    OUTPUT_PATH="dist/${GOOS}_${GOARCH}/${OUTPUT_NAME}"

    echo "构建 $GOOS/$GOARCH..."

    # 创建输出目录
    mkdir -p "dist/${GOOS}_${GOARCH}"

    # 构建
    GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -o "$OUTPUT_PATH" .

    if [ $? -eq 0 ]; then
        echo "✓ 成功: $OUTPUT_PATH"
    else
        echo "✗ 失败: $GOOS/$GOARCH"
    fi
done

echo ""
echo "构建完成！产物位于 dist/ 目录"
echo ""
echo "部署说明："
echo "1. 将对应平台的可执行文件复制到插件目录"
echo "2. 确保 plugin.yaml 中的 platform 配置正确"
echo "3. 重启主程序或执行 reload 命令加载插件"
