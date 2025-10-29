# Installer 快速参考

## 文件概览

| 文件 | 行数 | 职责 | 主要函数 |
|------|------|------|----------|
| `install-in-docker.go` | ~200 | 主程序和流程控制 | `main()`, `getVersionInfo()` |
| `cmd.go` | ~90 | 命令执行 | `detectDiskMount()`, `detectArchitecture()`, `executeScript()` |
| `download.go` | ~150 | 下载和校验 | `downloadFile()`, `downloadFromURL()`, `verifySHA256()` |
| `extract.go` | ~90 | 解压和权限 | `extractTarGz()`, `setBinPermissions()` |

## 函数速查

### 命令执行 (cmd.go)
```go
detectDiskMount() (string, error)
// 检测硬盘挂载点，返回路径如 /mnt/media_rw/xxx

detectArchitecture() (string, error)
// 检测架构，返回 "arm64" 或 "x86_64"

executeScript(scriptPath string) error
// 执行脚本并实时输出日志

streamOutput(reader io.Reader)
// 流式输出日志到控制台
```

### 下载模块 (download.go)
```go
downloadFile(destPath, filename, expectedSHA256 string) error
// 从 CDN/服务器下载文件并验证 SHA256
// 自动重试多个源

downloadFromURL(url, destPath string) error
// 从指定 URL 下载，带进度显示

verifySHA256(filePath, expectedHash string) error
// 验证文件 SHA256 哈希
```

### 解压模块 (extract.go)
```go
extractTarGz(tarGzPath, destDir string) error
// 解压 tar.gz 文件到目标目录

setBinPermissions(binDir string) error
// 设置目录中所有文件的可执行权限 (0755)
```

### 主程序 (install-in-docker.go)
```go
main()
// 安装流程入口

getVersionInfo(tmpDir string) (*VersionInfo, error)
// 下载并解析 version.txt
```

## 安装流程

```
┌─────────────────────────────────────┐
│ 1. 检测硬盘挂载                     │
│    detectDiskMount()                │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ 2. 获取版本信息                     │
│    getVersionInfo()                 │
│    ├─ detectArchitecture()          │
│    └─ downloadFile(version.txt)     │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ 3. 下载安装文件                     │
│    downloadFile(docker-*.tar.gz)    │
│    downloadFile(bin-*.tar.gz)       │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ 4. 解压文件                         │
│    extractTarGz() x2                │
│    setBinPermissions()              │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ 5. 执行部署脚本                     │
│    executeScript()                  │
│    └─ streamOutput() x2             │
└─────────────────────────────────────┘
```

## 常用操作

### 编译
```bash
# arm64
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
  go build -ldflags="-s -w" -o install-docker-arm64

# x86_64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w" -o install-docker-x86_64
```

### 测试编译
```bash
cd installer
go build -o /tmp/test-installer
./tmp/test-installer  # 在 Android 设备上测试
```

### 代码检查
```bash
go fmt ./...           # 格式化代码
go vet ./...           # 静态分析
golangci-lint run      # 代码检查（需安装）
```

## 错误处理

| 模块 | 策略 | 示例 |
|------|------|------|
| cmd | 命令失败返回错误 | 硬盘未挂载 → 退出 |
| download | 多源重试 | CDN 失败 → 尝试服务器 |
| download | SHA256 失败清理 | 校验失败 → 删除文件 → 重试 |
| extract | 立即返回 | 解压失败 → 退出 |
| extract | 忽略符号链接错误 | symlink 失败 → 继续 |

## 配置常量

### download.go
```go
cdnURL    = "https://fw.kspeeder.com/binary/docker-for-android"
serverURL = "https://fw.koolcenter.com/binary/docker-for-android"
```

### install-in-docker.go
```go
dockerRoot  = "/data/local/docker"
binDir      = "/data/local/docker/bin"
versionFile = "version.txt"
```

## 调试技巧

### 添加详细日志
```go
// 在函数开头添加
fmt.Printf("[DEBUG] 函数名: 参数=%v\n", param)
```

### 跳过某步骤测试
```go
// 注释掉某个步骤
// if err := downloadFile(...); err != nil {
//     return err
// }
fmt.Println("⚠ 跳过下载步骤（测试模式）")
```

### 使用本地文件测试
```go
// 修改 downloadFile 直接复制本地文件
func downloadFile(destPath, filename, expectedSHA256 string) error {
    // 测试：从本地复制而不是下载
    return os.Link("/path/to/test/"+filename, destPath)
}
```

## 扩展示例

### 添加新架构
在 `cmd.go` 的 `detectArchitecture()` 中添加：
```go
case "riscv64":
    return "riscv64", nil
```

### 添加新下载源
在 `download.go` 的 `downloadFile()` 中添加：
```go
urls := []string{
    fmt.Sprintf("%s/%s", cdnURL, filename),
    fmt.Sprintf("%s/%s", serverURL, filename),
    fmt.Sprintf("%s/%s", mirrorURL, filename),  // 新增
}
```

### 自定义进度显示
在 `download.go` 的 `downloadFromURL()` 中修改：
```go
fmt.Printf("   ⏬ %.1f%% [%s] %dMB/%dMB\r",
    progress,
    progressBar(progress),  // 自定义进度条
    written/(1024*1024),
    contentLength/(1024*1024))
```

## 资源链接

- [ARCHITECTURE.md](ARCHITECTURE.md) - 详细架构文档
- [README.md](README.md) - 使用说明
- [TESTING.md](TESTING.md) - 测试指南
- [Go 标准库文档](https://pkg.go.dev/std)
