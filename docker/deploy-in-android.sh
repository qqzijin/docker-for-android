#!/bin/sh

set -e

DOCKER_ROOT=/data/local/docker
cd "$DOCKER_ROOT"

echo "=========================================="
echo "Docker for Android - Deployment Script"
echo "=========================================="
echo ""

# Step 1: 检测硬盘挂载点并设置 DISK_ROOT
echo "[1/4] 检测硬盘挂载点..."
NVME=$(mount | grep -F '/dev/block/vold/public:259,1 on /mnt/media_rw/' | grep -F ' type ext4 ' | grep -oE '/mnt/media_rw/[^ ]+')

if [ -n "$NVME" ]; then
    # 设置硬盘根目录
    DISK_ROOT="$NVME"
    echo "✓ 检测到硬盘挂载点: $NVME"
    echo "✓ 硬盘根目录: $DISK_ROOT"
        
    # 从 DISK_ROOT 派生的路径会自动在 docker.env 中定义
    # 这里只需要创建基础目录
    mkdir -p "$DISK_ROOT/opt/dockerd/docker"
    mkdir -p "$DISK_ROOT/Cache/Kspeeder"
    mkdir -p "$DISK_ROOT/Configs/DPanel"

    # 创建基础目录结构
    touch "$DISK_ROOT/opt/.nomedia"
    touch "$DISK_ROOT/Cache/.nomedia"
    touch "$DISK_ROOT/Configs/.nomedia"
    
    echo "✓ 目录结构已创建："
    echo "  - Docker 数据: \$DISK_ROOT/dockerd/docker"
    echo "  - 缓存目录: \$DISK_ROOT/Cache/Kspeeder"
    echo "  - DPanel 配置: \$DISK_ROOT/Configs/DPanel"
else
    echo "✗ 错误: 未检测到硬盘挂载点"
    echo "✗ Docker 需要 ext4 格式的外置硬盘才能运行"
    echo "✗ 请确保已接入并格式化硬盘后再运行此脚本"
    exit 1
fi
echo ""

# Step 2: 更新 docker.env 文件
echo "[2/4] 配置环境变量..."
if [ -f "$DOCKER_ROOT/docker.env" ]; then
    # 备份原文件
    cp "$DOCKER_ROOT/docker.env" "$DOCKER_ROOT/docker.env.bak"
    
    # 只需要更新 DISK_ROOT，其他变量会自动派生
    sed -i "s|^export DISK_ROOT=.*|export DISK_ROOT=$DISK_ROOT|g" "$DOCKER_ROOT/docker.env"
    
    echo "✓ 已更新 docker.env:"
    echo "  - DISK_ROOT=$DISK_ROOT"
    echo "  - DOCKER_DATA_ROOT=\${DISK_ROOT}/dockerd/docker (自动派生)"
    echo "  - DISK_CACHE=\${DISK_ROOT}/Cache (自动派生)"
else
    echo "✗ 错误: docker.env 文件不存在"
    exit 1
fi
echo ""

# Step 3: 更新 kspeeder 配置（通过环境变量）
echo "[3/4] 配置 kspeeder..."
if [ -f "$DOCKER_ROOT/etc/kspeeder.conf" ]; then
    # 更新环境变量配置
    if grep -q "environment=" "$DOCKER_ROOT/etc/kspeeder.conf"; then
        # 备份原文件
        cp "$DOCKER_ROOT/etc/kspeeder.conf" "$DOCKER_ROOT/etc/kspeeder.conf.bak"
        sed -i "s|environment=DISK_CACHE=.*|environment=DISK_CACHE=\"$DISK_ROOT/Cache\"|g" "$DOCKER_ROOT/etc/kspeeder.conf"
        echo "✓ 已更新 kspeeder 环境变量配置"
    else
        echo "✓ kspeeder 将使用 docker.env 中的环境变量"
    fi
    echo "✓ kspeeder 缓存路径: \$DISK_CACHE/Kspeeder"
else
    echo "⚠ kspeeder.conf 文件不存在，跳过配置"
fi
echo ""

# Step 4: 启动 Docker 并部署 DPanel
echo "[4/4] 启动 Docker 服务..."

# 加载环境变量
. "$DOCKER_ROOT/docker.env"

# 启动 supervisor（会启动 dockerd 和 kspeeder）
if [ -f "$DOCKER_ROOT/start.sh" ]; then
    echo "✓ 正在启动 Docker 服务..."
    nohup sh "$DOCKER_ROOT/start.sh" > /dev/null 2>&1 &
    DOCKERD_PID=$!
    echo "✓ Docker 服务已启动 (PID: $DOCKERD_PID)"
else
    echo "✗ 错误: start.sh 文件不存在"
    exit 1
fi

# 等待 Docker 就绪
echo "⏳ 等待 Docker 服务就绪..."
MAX_WAIT=60
WAIT_COUNT=0
while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    if docker info > /dev/null 2>&1; then
        echo "✓ Docker 服务已就绪"
        break
    fi
    sleep 2
    WAIT_COUNT=$((WAIT_COUNT + 2))
    echo "   等待中... ($WAIT_COUNT/$MAX_WAIT 秒)"
done

if [ $WAIT_COUNT -ge $MAX_WAIT ]; then
    echo "✗ Docker 服务启动超时"
    exit 1
fi
echo ""

# 部署 DPanel
echo "=========================================="
echo "部署 DPanel 容器管理面板"
echo "=========================================="

# DPanel 配置目录从 DISK_ROOT 派生
DPANEL_CONFIG_DIR="$DISK_ROOT/Configs/DPanel"
echo "✓ DPanel 配置目录: $DPANEL_CONFIG_DIR"

echo "⏳ 拉取 DPanel 镜像..."
if docker pull dpanel/dpanel:latest; then
    echo "✓ DPanel 镜像拉取成功"
else
    echo "⚠ DPanel 镜像拉取失败，请检查网络连接"
fi

echo "⏳ 启动 DPanel 容器..."
docker run -d \
    --name dpanel \
    --restart always \
    -p 8080:8080 \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v "$DPANEL_CONFIG_DIR:/dpanel" \
    dpanel/dpanel:latest

if [ $? -eq 0 ]; then
    echo "✓ DPanel 已成功部署"
    echo ""
    echo "=========================================="
    echo "部署完成！"
    echo "=========================================="
    echo "硬盘根目录: $DISK_ROOT"
    echo "Docker 数据: \$DISK_ROOT/dockerd/docker"
    echo "缓存目录: \$DISK_ROOT/Cache"
    echo "DPanel 配置: $DPANEL_CONFIG_DIR"
    echo "DPanel 访问地址: http://<设备IP>:8080"
    echo "=========================================="
else
    echo "✗ DPanel 部署失败"
    exit 1
fi
