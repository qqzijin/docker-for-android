# 停止 Supervisord 服务功能实现

## 概述

在安装程序中新增了 `stopSupervisord()` 功能，用于在解压新文件之前停止正在运行的 supervisord 服务，避免升级安装时因服务冲突导致的启动失败。

## 实现位置

- **主函数调用**：`installer/install-in-docker.go`
- **功能实现**：`installer/cmd.go`
- **单元测试**：`installer/cmd_test.go`

## 功能流程

### 1. 主安装流程集成（install-in-docker.go）

在下载完成后、解压文件之前，增加新的步骤 [3.5/5]：

```go
// Step 3.5: 停止正在运行的服务
fmt.Println("[3.5/5] 停止现有服务...")
if err := stopSupervisord(); err != nil {
    fmt.Printf("✗ 错误: 停止 supervisord 服务失败: %v\n", err)
    os.Exit(1)
}
fmt.Println()
```

### 2. 核心功能实现（cmd.go）

#### `stopSupervisord()` - 主函数

**功能**：停止正在运行的 supervisord 服务

**执行步骤**：

1. **检查文件存在性**
   ```go
   supervisordPath := "/data/local/docker/bin/supervisord"
   if _, err := os.Stat(supervisordPath); os.IsNotExist(err) {
       // 文件不存在，说明是首次安装
       fmt.Println("✓ 未检测到已安装的 supervisord，跳过停止服务")
       return nil
   }
   ```

2. **优雅停止服务**
   ```go
   cmd := exec.Command(supervisordPath, "ctl", "stop", "all")
   output, err := cmd.CombinedOutput()
   ```
   - 执行 `supervisorctl stop all` 命令
   - 如果命令失败，记录警告但继续执行（可能 supervisord 未运行）

3. **等待进程退出**
   ```go
   time.Sleep(2 * time.Second)
   ```
   - 给予进程 2 秒时间完成正常退出流程

4. **检查残留进程**
   ```go
   pids, err := findSupervisordProcesses()
   if len(pids) == 0 {
       fmt.Println("✓ 未检测到 supervisord 进程")
       return nil
   }
   ```

5. **强制终止进程**
   ```go
   for _, pid := range pids {
       fmt.Printf("  终止进程 PID: %d\n", pid)
       if err := killProcess(pid); err != nil {
           fmt.Printf("  ⚠ 警告: 终止进程 %d 失败: %v\n", pid, err)
       }
   }
   ```
   - 使用 `kill -9` 强制终止所有 supervisord 进程

6. **最终验证**
   ```go
   time.Sleep(1 * time.Second)
   pids, _ = findSupervisordProcesses()
   if len(pids) > 0 {
       return fmt.Errorf("无法终止 supervisord 进程: %v", pids)
   }
   ```

#### `findSupervisordProcesses()` - 查找进程

**功能**：查找所有正在运行的 supervisord 进程

**实现**：
```go
func findSupervisordProcesses() ([]int, error) {
    cmd := exec.Command("ps", "-ef")
    output, err := cmd.Output()
    // ...
    
    for _, line := range lines {
        if strings.Contains(line, "supervisord") && !strings.Contains(line, "grep") {
            fields := strings.Fields(line)
            if len(fields) >= 2 {
                pid, err := strconv.Atoi(fields[1])
                if err == nil {
                    pids = append(pids, pid)
                }
            }
        }
    }
    return pids, nil
}
```

**特点**：
- 使用 `ps -ef` 列出所有进程
- 过滤包含 "supervisord" 的进程
- 排除 grep 进程本身
- 解析 PID（第二列）

#### `killProcess()` - 终止进程

**功能**：使用 `kill -9` 强制终止指定 PID 的进程

```go
func killProcess(pid int) error {
    cmd := exec.Command("kill", "-9", strconv.Itoa(pid))
    return cmd.Run()
}
```

## 错误处理策略

### 1. 首次安装场景
- **情况**：`/data/local/docker/bin/supervisord` 不存在
- **处理**：直接返回 nil，跳过停止服务步骤
- **输出**：`✓ 未检测到已安装的 supervisord，跳过停止服务`

### 2. supervisorctl 命令失败
- **情况**：supervisord 可能未运行或命令执行失败
- **处理**：记录警告但继续执行，不中断安装流程
- **输出**：`⚠ supervisorctl stop all 执行警告`

### 3. 进程查找失败
- **情况**：`ps` 命令执行失败
- **处理**：返回错误，中断安装
- **输出**：`✗ 错误: 停止 supervisord 服务失败`

### 4. 无法终止进程
- **情况**：kill 后进程仍然存在
- **处理**：返回错误，中断安装
- **输出**：`✗ 错误: 无法终止 supervisord 进程`

## 用户体验

### 首次安装输出示例
```
[3.5/5] 停止现有服务...
✓ 未检测到已安装的 supervisord，跳过停止服务
```

### 升级安装输出示例（正常情况）
```
[3.5/5] 停止现有服务...
⏳ 检测到已安装的 supervisord，正在停止服务...
✓ supervisord 服务已停止
  docker-proxy: stopped
  dockerd: stopped
  ...
⏳ 检查 supervisord 进程...
✓ 未检测到 supervisord 进程
```

### 升级安装输出示例（需强制终止）
```
[3.5/5] 停止现有服务...
⏳ 检测到已安装的 supervisord，正在停止服务...
✓ supervisord 服务已停止
⏳ 检查 supervisord 进程...
⚠ 检测到 1 个 supervisord 进程仍在运行，正在强制终止...
  终止进程 PID: 12345
✓ 所有 supervisord 进程已终止
```

## 测试

### 单元测试（cmd_test.go）

实现了以下测试用例：

1. **TestFindSupervisordProcesses_NoProcesses**
   - 测试在没有 supervisord 进程时的行为
   - 验证返回的 PID 列表有效性

2. **TestKillProcess_InvalidPID**
   - 测试 kill 不存在的进程
   - 验证错误处理

3. **TestStopSupervisord_NoInstallation**
   - 测试首次安装场景（supervisord 不存在）

4. **TestParsePIDFromPSOutput**
   - 测试从 ps 输出解析 PID 的逻辑
   - 验证各种格式的进程行

5. **TestStopSupervisordWorkflow**
   - 文档性测试，验证整体工作流程

6. **BenchmarkFindSupervisordProcesses**
   - 性能基准测试

### 运行测试

```bash
cd installer
go test -v -run TestFindSupervisord
go test -v -run TestKillProcess
go test -v -run TestParsePID
go test -bench=BenchmarkFindSupervisord
```

## 技术细节

### 依赖的包
```go
import (
    "os"
    "os/exec"
    "strconv"
    "strings"
    "time"
)
```

### 时间等待说明
- **第一次等待（2 秒）**：在执行 `supervisorctl stop all` 后，给予进程足够时间完成正常退出流程
- **第二次等待（1 秒）**：在 `kill -9` 后，等待系统完成进程清理

### 进程识别逻辑
- 使用 `ps -ef` 而不是 `ps aux`，因为在 Android 系统上更通用
- 排除包含 "grep" 的行，避免将 grep 进程本身识别为 supervisord
- PID 通常是第二列（索引为 1）

## 兼容性

- ✅ **arm64 架构**：完全支持
- ✅ **x86_64 架构**：完全支持
- ✅ **Android 系统**：针对 Android 系统优化
- ✅ **首次安装**：自动检测并跳过停止服务
- ✅ **升级安装**：自动停止现有服务后安装

## 安全性考虑

1. **优雅停止优先**：首先尝试使用 `supervisorctl stop all` 优雅停止服务
2. **强制终止后备**：只有在优雅停止失败时才使用 `kill -9`
3. **进程验证**：在终止前后都会验证进程状态
4. **错误传播**：关键错误会中断安装流程，避免数据损坏

## 未来改进建议

1. **可配置的等待时间**：允许通过环境变量配置等待时间
2. **更详细的日志**：记录每个被终止进程的详细信息
3. **进程树终止**：考虑终止 supervisord 的所有子进程
4. **重试机制**：在 kill 失败时增加重试逻辑

## 相关文档

- [ARCHITECTURE.md](ARCHITECTURE.md) - 整体代码架构说明
- [TESTING.md](TESTING.md) - 测试文档
- [README.md](README.md) - 项目说明

## 变更历史

- **2025-10-29**：首次实现停止 supervisord 功能
  - 增加 `stopSupervisord()` 主函数
  - 增加 `findSupervisordProcesses()` 辅助函数
  - 增加 `killProcess()` 辅助函数
  - 集成到主安装流程
  - 添加单元测试
