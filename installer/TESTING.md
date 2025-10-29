# Installer 测试指南

## 单元测试

已经为可以在开发机器上运行的功能实现了完整的单元测试。

### 测试覆盖

#### 下载模块测试 (`download_test.go`)

✅ **基础功能测试**
- `TestDownloadFromURL` - HTTP 下载功能
- `TestDownloadFromURL_HTTPError` - HTTP 错误处理
- `TestVerifySHA256` - SHA256 校验
- `TestVerifySHA256_FileNotExist` - 文件不存在错误处理

✅ **高级功能测试**
- `TestDownloadFile_WithRetry` - 多源重试机制
- `TestDownloadFile_SHA256Mismatch` - SHA256 不匹配处理
- `TestGetVersionInfo_Parse` - 版本文件解析
- `TestGetVersionInfo_MissingFields` - 缺少必需字段验证

✅ **性能和压力测试**
- `TestDownloadWithProgress` - 带进度显示的大文件下载（5MB）
- `TestLargeFileDownload` - 大文件下载测试（50MB）
- `TestStreamOutput` - 日志流式输出

✅ **性能基准测试**
- `BenchmarkDownloadFromURL` - 下载性能基准
- `BenchmarkVerifySHA256` - SHA256 校验性能基准

#### 解压模块测试 (`extract_test.go`)

✅ **基础功能测试**
- `TestExtractTarGz` - tar.gz 解压功能
- `TestExtractTarGz_EmptyArchive` - 空归档处理
- `TestExtractTarGz_NestedDirectories` - 嵌套目录解压
- `TestExtractTarGz_FileNotExist` - 文件不存在错误处理
- `TestExtractTarGz_InvalidGzip` - 无效 gzip 文件处理

✅ **权限测试**
- `TestSetBinPermissions` - 设置二进制文件权限
- `TestSetBinPermissions_WithSubdirectories` - 带子目录的权限设置
- `TestSetBinPermissions_EmptyDirectory` - 空目录处理
- `TestSetBinPermissions_DirectoryNotExist` - 目录不存在错误处理
- `TestExtractTarGz_PreservePermissions` - 保持文件权限

✅ **性能测试**
- `TestExtractTarGz_LargeFile` - 大文件解压测试（10MB）
- `BenchmarkExtractTarGz` - 解压性能基准

### 未测试功能（需要 Android 环境）

以下功能需要在 Android 设备上测试：

⏸️ **命令执行模块** (`cmd.go`)
- `detectDiskMount()` - 需要 Android 的 mount 命令和硬盘
- `detectArchitecture()` - 可测试但在当前系统上结果不同
- `executeScript()` - 需要 Android 环境和脚本
- `streamOutput()` - 已通过模拟测试

⏸️ **完整安装流程**
- `main()` - 需要完整的 Android 环境
- `getVersionInfo()` - 部分逻辑已测试（解析部分）

## 运行测试

### 运行所有测试

```bash
cd installer
go test -v
```

### 运行特定测试

```bash
# 测试下载功能
go test -v -run TestDownload

# 测试解压功能
go test -v -run TestExtract

# 测试 SHA256 验证
go test -v -run TestVerifySHA256
```

### 跳过耗时测试

```bash
go test -v -short
```

### 运行性能基准测试

```bash
# 运行所有基准测试
go test -bench=. -benchmem

# 运行特定基准测试
go test -bench=BenchmarkDownload -benchmem
go test -bench=BenchmarkVerifySHA256 -benchmem
go test -bench=BenchmarkExtract -benchmem
```

### 查看测试覆盖率

```bash
# 生成覆盖率报告
go test -coverprofile=coverage.out

# 查看覆盖率
go tool cover -func=coverage.out

# 生成 HTML 报告
go tool cover -html=coverage.out
```

## 测试结果示例

### 成功输出

```
=== RUN   TestDownloadFromURL
   进度: 100.0% (0/0 MB)
--- PASS: TestDownloadFromURL (0.00s)
=== RUN   TestVerifySHA256
--- PASS: TestVerifySHA256 (0.00s)
=== RUN   TestExtractTarGz
--- PASS: TestExtractTarGz (0.00s)
=== RUN   TestSetBinPermissions
--- PASS: TestSetBinPermissions (0.00s)
...
PASS
ok      github.com/linkease/docker-for-android/installer        0.445s
```

### 性能基准示例

```
BenchmarkDownloadFromURL-8     100    12345678 ns/op    1234567 B/op    123 allocs/op
BenchmarkVerifySHA256-8        50     23456789 ns/op    2345678 B/op    234 allocs/op
BenchmarkExtractTarGz-8        30     34567890 ns/op    3456789 B/op    345 allocs/op
```

## 测试场景说明

### 下载测试

1. **正常下载** - 模拟从 HTTP 服务器下载文件
2. **进度显示** - 验证下载进度正确显示
3. **错误处理** - 测试 404、500 等 HTTP 错误
4. **SHA256 校验** - 验证文件完整性检查
5. **大文件下载** - 测试 50MB 文件下载性能

### 解压测试

1. **基础解压** - 解压包含多个文件的归档
2. **嵌套目录** - 正确处理深层目录结构
3. **权限保持** - 保持原始文件权限（0755, 0644, 0444）
4. **空归档** - 处理空的 tar.gz 文件
5. **错误处理** - 无效文件格式、文件不存在等

### 版本信息测试

1. **正确解析** - 解析 version.txt 各字段
2. **字段验证** - 检测缺少必需字段
3. **注释处理** - 忽略注释行和空行
4. **多架构** - 正确识别不同架构的 SHA256

## 编译测试

### 本地编译

```bash
cd installer
go build -o test-installer
```

### 交叉编译测试

```bash
# 测试 arm64 编译
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o test-arm64

# 测试 x86_64 编译
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o test-x86_64
```

## 本地测试建议

虽然不能在本地完整运行安装流程，但可以测试各个模块：

### 1. 测试下载功能

创建一个简单的 HTTP 服务器提供测试文件：

```bash
# 启动测试服务器
python3 -m http.server 8000

# 修改代码临时使用本地服务器
# 在 download.go 中临时修改 cdnURL
```

### 2. 测试解压功能

使用现有的 tar.gz 文件进行测试：

```bash
# 创建测试 tar.gz
tar -czf test.tar.gz docker/

# 运行测试
go test -v -run TestExtract
```

### 3. 测试版本解析

创建测试 version.txt：

```bash
cat > test-version.txt << EOF
VERSION=28.0.1.10
DOCKER_SHA256=abc123
BIN_ARM64_SHA256=def456
EOF

# 运行测试
go test -v -run TestGetVersionInfo
```

## 在 Android 设备上测试

### 准备工作

1. 确保设备已 root
2. 确保已接入并格式化 ext4 外置硬盘
3. 确保网络连接正常

### 推送测试

```bash
# 推送到设备
adb push release/install-docker-arm64 /data/local/tmp/install-docker
adb shell chmod +x /data/local/tmp/install-docker

# 进入设备
adb shell
su

# 测试硬盘检测（不执行安装）
mount | grep -F '/dev/block/vold/public:259,1 on /mnt/media_rw/' | grep -F ' type ext4 ' | grep -oE '/mnt/media_rw/[^ ]+'
# 应该输出类似: /mnt/media_rw/xxx

# 执行安装
/data/local/tmp/install-docker
```

## 持续集成建议

### GitHub Actions 示例

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: |
          cd installer
          go test -v -coverprofile=coverage.out
          go tool cover -func=coverage.out
      
      - name: Run benchmarks
        run: |
          cd installer
          go test -bench=. -benchmem
      
      - name: Test build
        run: |
          cd installer
          CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build
          CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
```

## 故障排查

### 测试失败

1. **检查依赖** - 确保 Go 版本 >= 1.21
2. **清理缓存** - `go clean -testcache`
3. **查看详细日志** - `go test -v`
4. **单独运行** - `go test -v -run TestName`

### 性能问题

1. **跳过大文件测试** - `go test -short`
2. **调整测试大小** - 修改测试中的文件大小
3. **检查磁盘空间** - 确保 /tmp 有足够空间

### 权限问题

某些测试需要文件权限支持，如果在 Windows 上运行可能有差异。建议在 Linux/macOS 上运行完整测试。

## 测试统计

当前测试覆盖：

- ✅ 下载模块：12 个测试 + 2 个基准测试
- ✅ 解压模块：11 个测试 + 1 个基准测试
- ✅ 总计：23 个单元测试，3 个基准测试
- ⏸️ Android 相关功能需要设备测试

预期覆盖率：~70-80%（排除 Android 特定代码）


### 1. 编译 arm64 版本

```bash
cd installer
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o install-docker-arm64 install-in-docker.go
```

检查编译结果：
```bash
file install-docker-arm64
# 应该输出: install-docker-arm64: ELF 64-bit LSB executable, ARM aarch64...
```

### 2. 编译 x86_64 版本

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o install-docker-x86_64 install-in-docker.go
```

### 3. 使用 Makefile 编译

```bash
cd ..
make installer
```

## 代码审查检查点

### 1. 版本文件解析

installer 期望的 version.txt 格式：

```
VERSION=28.0.1.10
DOCKER_SHA256=<sha256>
BIN_ARM64_SHA256=<sha256>
BIN_X86_64_SHA256=<sha256>
```

### 2. 下载 URL 构造

- CDN: `https://fw.kspeeder.com/binary/docker-for-android/<filename>`
- 服务器: `https://fw.koolcenter.com/binary/docker-for-android/<filename>`

文件名格式：
- `docker-28.0.1.10.tar.gz`
- `docker-for-android-bin-28.0.1.10-arm64.tar.gz`
- `docker-for-android-bin-28.0.1.10-x86_64.tar.gz`

### 3. 解压目标路径

- Docker 包解压到: `/data/local/` (会创建 `/data/local/docker/`)
- 二进制包解压到: `/data/local/docker/bin/`

### 4. 硬盘检测命令

```bash
mount | grep -F '/dev/block/vold/public:259,1 on /mnt/media_rw/' | grep -F ' type ext4 ' | grep -oE '/mnt/media_rw/[^ ]+'
```

## 在 Android 设备上测试

### 准备工作

1. 确保设备已 root
2. 确保已接入并格式化 ext4 外置硬盘
3. 确保网络连接正常

### 推送测试

```bash
# 推送到设备
adb push release/install-docker-arm64 /data/local/tmp/install-docker
adb shell chmod +x /data/local/tmp/install-docker

# 进入设备
adb shell
su

# 测试硬盘检测（不执行安装）
mount | grep -F '/dev/block/vold/public:259,1 on /mnt/media_rw/' | grep -F ' type ext4 ' | grep -oE '/mnt/media_rw/[^ ]+'
# 应该输出类似: /mnt/media_rw/xxx

# 执行安装
/data/local/tmp/install-docker
```

### 预期输出

```
==========================================
Docker for Android - Installer
==========================================

[1/5] 检测硬盘挂载点...
✓ 检测到硬盘挂载点: /mnt/media_rw/xxx
✓ 临时目录: /mnt/media_rw/xxx/Cache/installer

[2/5] 获取版本信息...
   尝试从CDN下载...
   进度: 100.0% (0/0 MB)
   ✓ SHA256 验证通过
✓ 版本: 28.0.1.10
✓ 架构: arm64

[3/5] 下载安装文件...
⏳ 下载 docker-28.0.1.10.tar.gz...
   尝试从CDN下载...
   进度: 100.0% (xx/xx MB)
   ✓ SHA256 验证通过
✓ docker-28.0.1.10.tar.gz 下载完成
⏳ 下载 docker-for-android-bin-28.0.1.10-arm64.tar.gz...
   尝试从CDN下载...
   进度: 100.0% (xx/xx MB)
   ✓ SHA256 验证通过
✓ docker-for-android-bin-28.0.1.10-arm64.tar.gz 下载完成

[4/5] 解压安装文件...
⏳ 解压 docker-28.0.1.10.tar.gz 到 /data/local...
✓ docker-28.0.1.10.tar.gz 解压完成
⏳ 解压 docker-for-android-bin-28.0.1.10-arm64.tar.gz 到 /data/local/docker/bin...
✓ docker-for-android-bin-28.0.1.10-arm64.tar.gz 解压完成

[5/5] 执行部署脚本...
==========================================
Docker for Android - Deployment Script
==========================================
...

==========================================
安装完成！
==========================================
```

## 错误场景测试

### 1. 无硬盘挂载

预期输出：
```
[1/5] 检测硬盘挂载点...
✗ 错误: 未检测到硬盘挂载点
✗ Docker 需要 ext4 格式的外置硬盘才能运行
✗ 请确保已接入并格式化硬盘后再运行此程序
```

### 2. 网络不可用

预期输出：
```
[3/5] 下载安装文件...
⏳ 下载 docker-28.0.1.10.tar.gz...
   尝试从CDN下载...
   ✗ 下载失败: ...
   尝试从服务器下载...
   ✗ 下载失败: ...
✗ 错误: 所有下载源均失败: ...
```

### 3. SHA256 校验失败

预期输出：
```
   尝试从CDN下载...
   进度: 100.0% (xx/xx MB)
   ✗ SHA256 验证失败: SHA256 不匹配 (期望: xxx, 实际: yyy)
   尝试从服务器下载...
```

## 调试技巧

### 1. 查看详细日志

installer 会实时输出所有操作日志，包括：
- 每个步骤的进度
- 下载进度（带百分比和 MB 数）
- deploy-in-android.sh 脚本的实时输出

### 2. 手动验证文件

```bash
# 检查临时文件
ls -lh /mnt/media_rw/xxx/Cache/installer/

# 检查安装结果
ls -lh /data/local/docker/
ls -lh /data/local/docker/bin/

# 检查权限
ls -la /data/local/docker/bin/docker*
```

### 3. 手动测试部署脚本

```bash
cd /data/local/docker
sh -x deploy-in-android.sh  # -x 参数显示详细执行过程
```

## 清理测试环境

```bash
# 停止 Docker
docker stop $(docker ps -aq)
supervisorctl stop all

# 删除安装文件
rm -rf /data/local/docker
rm /data/local/tmp/install-docker

# 清理硬盘上的数据（慎重！）
# rm -rf /mnt/media_rw/xxx/opt/dockerd
# rm -rf /mnt/media_rw/xxx/Cache
# rm -rf /mnt/media_rw/xxx/Configs
```

## 持续集成建议

1. 在 CI 中编译 arm64 和 x86_64 版本
2. 计算编译产物的 SHA256
3. 自动上传到服务器
4. 更新 version.txt 文件
5. 在真实设备上进行冒烟测试
