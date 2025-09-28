#!/bin/bash

# UI构建脚本
set -e

echo "Building ClickHouse Sinker UI..."

# 创建dist目录
mkdir -p mvc/dist

# 复制单一HTML文件（包含所有CSS和JS）
cp mvc/static/index.html mvc/dist/

# 检查是否安装了HTML压缩工具
if command -v html-minifier &> /dev/null; then
    echo "Compressing HTML..."
    html-minifier --collapse-whitespace --remove-comments --minify-css --minify-js mvc/dist/index.html -o mvc/dist/index.html 2>/dev/null || echo "HTML compression failed, using original"
else
    echo "html-minifier not found, skipping HTML optimization..."
fi

# 生成文件哈希（用于缓存控制）
if command -v sha256sum &> /dev/null; then
    echo "Generating file hashes..."
    cd mvc/dist
    sha256sum *.html > checksums.txt 2>/dev/null || echo "Hash generation failed"
    cd - > /dev/null
fi

echo "UI build completed successfully!"
echo "Files generated in mvc/dist/:"
ls -la mvc/dist/