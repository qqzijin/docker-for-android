# Installer 代码结构说明

## 文件组织

代码已重构为清晰的模块化结构，每个文件负责特定的功能：

### 1. `install-in-docker.go` - 主程序

**职责：** 程序入口和流程控制

**主要内容：**
- `main()` - 主函数，控制整个安装流程
- `VersionInfo` - 版本信息结构体
- `getVersionInfo()` - 解析版本文件
- 常量定义（安装路径、版本文件名等）

**流程：**
1. 检测硬盘挂载
2. 获取版本信息
3. 下载文件
4. 解压文件
5. 执行部署脚本

### 2. `cmd.go` - 命令行执行模块

**职责：** 所有与系统命令执行相关的功能

**主要函数：**
- `detectDiskMount()` - 检测硬盘挂载点
  - 执行 `mount` 命令
  - 过滤 ext4 格式的硬盘
  - 返回挂载路径

- `detectArchitecture()` - 检测系统架构
  - 执行 `uname -m` 命令
  - 识别 arm64 或 x86_64
  - 处理架构别名

- `executeScript()` - 执行 shell 脚本
  - 启动脚本进程
  - 捕获标准输出和标准错误
  - 实时输出日志

- `streamOutput()` - 流式输出日志
  - 逐行读取输出
  - 实时打印到控制台

### 3. `download.go` - 下载模块

**职责：** 文件下载和校验

**常量：**
- `cdnURL` - CDN 地址
- `serverURL` - 源服务器地址

**主要函数：**
- `downloadFile()` - 下载文件并验证
  - 尝试多个下载源（CDN → 服务器）
  - 自动重试和回退
  - SHA256 校验
  - 失败时清理临时文件

- `downloadFromURL()` - 从指定 URL 下载
  - HTTP 客户端配置（超时设置）
  - 进度显示（百分比和 MB）
  - 流式下载，节省内存
  - 错误处理

- `verifySHA256()` - 验证文件完整性
  - 计算文件 SHA256
  - 与期望值比对
  - 返回详细错误信息

### 4. `extract.go` - 解压模块

**职责：** tar.gz 文件解压和权限设置

**主要函数：**
- `extractTarGz()` - 解压 tar.gz 文件
  - 打开 gzip 压缩文件
  - 读取 tar 归档
  - 处理不同类型的文件（目录、普通文件、符号链接）
  - 保持文件权限
  - 自动创建父目录

- `setBinPermissions()` - 设置二进制文件权限
  - 遍历目录中的文件
  - 设置可执行权限（0755）
  - 跳过子目录

## 模块依赖关系

```
install-in-docker.go (主程序)
    ├── cmd.go (命令执行)
    │   ├── detectDiskMount()
    │   ├── detectArchitecture()
    │   ├── executeScript()
    │   └── streamOutput()
    │
    ├── download.go (下载)
    │   ├── downloadFile()
    │   ├── downloadFromURL()
    │   └── verifySHA256()
    │
    └── extract.go (解压)
        ├── extractTarGz()
        └── setBinPermissions()
```

## 调用流程

```
main()
  ├─> detectDiskMount()              [cmd.go]
  │
  ├─> getVersionInfo()                [install-in-docker.go]
  │   ├─> detectArchitecture()      [cmd.go]
  │   └─> downloadFile()            [download.go]
  │       ├─> downloadFromURL()     [download.go]
  │       └─> verifySHA256()        [download.go]
  │
  ├─> downloadFile() x2              [download.go]
  │   ├─> docker 包
  │   └─> 二进制包
  │
  ├─> extractTarGz() x2              [extract.go]
  │   ├─> 解压 docker 包
  │   └─> 解压二进制包
  │
  ├─> setBinPermissions()            [extract.go]
  │
  └─> executeScript()                [cmd.go]
      └─> streamOutput() x2          [cmd.go]
          ├─> stdout
          └─> stderr
```

## 错误处理策略

### 命令执行（cmd.go）
- 命令失败返回错误
- 输出为空视为错误
- 脚本非零退出码返回错误

### 下载（download.go）
- 尝试所有下载源
- SHA256 失败自动删除文件并重试
- 记录最后一个错误返回

### 解压（extract.go）
- 任何解压错误立即返回
- 符号链接错误被忽略（continue）

### 主程序（install-in-docker.go）
- 任何步骤失败打印错误并退出
- 退出码 1 表示失败

## 代码特点

### 1. 模块化设计
- 每个文件职责单一
- 函数命名清晰
- 易于测试和维护

### 2. 可扩展性
- 新增下载源：修改 `download.go` 的 URL 列表
- 新增架构支持：修改 `cmd.go` 的 `detectArchitecture()`
- 新增命令执行：在 `cmd.go` 中添加函数

### 3. 用户友好
- 清晰的步骤提示
- 实时进度显示
- 详细的错误信息
- 彩色状态标记

### 4. 安全性
- SHA256 校验所有下载
- HTTPS 协议
- 失败时清理临时文件

## 编译方法

所有文件会被编译到一个二进制中：

```bash
# arm64
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o install-docker-arm64

# x86_64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o install-docker-x86_64
```

## 测试建议

### 单元测试
可以为每个模块编写独立的单元测试：

```bash
installer/
├── cmd_test.go
├── download_test.go
├── extract_test.go
└── install_test.go
```

### 集成测试
在真实 Android 设备上测试完整流程。

### Mock 测试
- Mock HTTP 响应测试下载
- Mock 命令输出测试检测
- Mock 文件系统测试解压

## 维护指南

### 添加新功能
1. 确定功能所属模块
2. 在对应文件中添加函数
3. 更新主流程调用
4. 更新文档

### 修复 Bug
1. 定位到具体模块和函数
2. 修复并添加测试
3. 验证不影响其他功能

### 性能优化
- 下载：考虑并行下载多个文件
- 解压：使用并发处理大量小文件
- 校验：使用增量校验大文件

## 依赖关系

**标准库：**
- `archive/tar` - tar 解压
- `compress/gzip` - gzip 解压
- `crypto/sha256` - 校验
- `encoding/hex` - 哈希编码
- `net/http` - HTTP 下载
- `os/exec` - 命令执行
- `bufio` - 行缓冲读取
- `io` - 流处理
- `fmt`, `os`, `path/filepath`, `strings`, `time` - 基础功能

**无外部依赖** - 可静态编译为单个二进制文件
