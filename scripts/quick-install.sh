#!/bin/bash

# 快速安装脚本 - 通过 adb 推送 installer 并执行

set -e

echo "=========================================="
echo "Docker for Android - Quick Install"
echo "=========================================="
echo ""

# 检查 adb 是否可用
if ! command -v adb &> /dev/null; then
    echo "✗ 错误: 找不到 adb 命令"
    echo "请确保 Android SDK Platform Tools 已安装并在 PATH 中"
    exit 1
fi

# 检查设备连接
echo "[1/4] 检查设备连接..."
DEVICES=$(adb devices | grep -v "List" | grep "device$" | wc -l)
if [ "$DEVICES" -eq 0 ]; then
    echo "✗ 错误: 未检测到 Android 设备"
    echo "请确保设备已连接并启用 USB 调试"
    exit 1
fi
echo "✓ 已连接 $DEVICES 个设备"
echo ""

# 检测设备架构
echo "[2/4] 检测设备架构..."
ARCH=$(adb shell uname -m | tr -d '\r\n')
echo "✓ 设备架构: $ARCH"

# 确定使用哪个 installer
INSTALLER=""
if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    INSTALLER="release/install-docker-arm64"
    TARGET_NAME="install-docker"
elif [ "$ARCH" = "x86_64" ] || [ "$ARCH" = "amd64" ]; then
    INSTALLER="release/install-docker-x86_64"
    TARGET_NAME="install-docker"
else
    echo "✗ 错误: 不支持的架构: $ARCH"
    exit 1
fi

# 检查 installer 文件是否存在
if [ ! -f "$INSTALLER" ]; then
    echo "✗ 错误: 找不到 installer 文件: $INSTALLER"
    echo "请先运行 'make installer' 构建安装程序"
    exit 1
fi
echo "✓ 使用安装程序: $INSTALLER"
echo ""

# 推送 installer 到设备
echo "[3/4] 推送安装程序到设备..."
adb push "$INSTALLER" "/data/local/tmp/$TARGET_NAME"
if [ $? -ne 0 ]; then
    echo "✗ 错误: 推送文件失败"
    exit 1
fi
echo "✓ 安装程序已推送到设备"

# 设置执行权限
adb shell "chmod +x /data/local/tmp/$TARGET_NAME"
echo "✓ 已设置执行权限"
echo ""

# 执行安装
echo "[4/4] 开始安装 Docker..."
echo "=========================================="
echo ""

# 提示用户需要 root 权限
echo "注意: 安装需要 root 权限"
echo "如果设备未 root，请手动执行以下命令："
echo ""
echo "  adb shell"
echo "  su"
echo "  /data/local/tmp/$TARGET_NAME"
echo ""
echo "按 Enter 继续自动安装（需要设备已 root）..."
read

# 尝试自动执行（如果设备已 root）
adb shell "su -c '/data/local/tmp/$TARGET_NAME'"

if [ $? -eq 0 ]; then
    echo ""
    echo "=========================================="
    echo "安装完成！"
    echo "=========================================="
    echo ""
    echo "DPanel 管理界面访问地址："
    echo "  http://<设备IP>:8080"
    echo ""
    echo "可以通过以下命令获取设备 IP："
    echo "  adb shell ip addr show wlan0"
else
    echo ""
    echo "✗ 自动安装失败"
    echo "请手动执行安装，或检查设备是否已 root"
fi
