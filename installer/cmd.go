package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"syscall"
)

// getDiskSize 获取指定路径的磁盘大小（KB）
func getDiskSize(path string) (uint64, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return 0, err
	}
	// 计算总大小（字节），然后转换为 KB
	// Blocks * BlockSize / 1024
	totalSizeKB := (stat.Blocks * uint64(stat.Bsize)) / 1024
	return totalSizeKB, nil
}

// detectDiskMount 检测硬盘挂载点
func detectDiskMount() (string, error) {
	// 纯 Go 实现：遍历 /mnt/media_rw 下 ext4 挂载点，选最大空间
	file, err := exec.Command("mount").Output()
	if err != nil {
		return "", fmt.Errorf("无法获取挂载信息: %v", err)
	}
	lines := strings.Split(string(file), "\n")
	var maxSize uint64
	var bestMount string
	for _, line := range lines {
		// 只处理 ext4 且挂载在 /mnt/media_rw 下的
		if !strings.Contains(line, " type ext4 ") || !strings.Contains(line, "/mnt/media_rw/") {
			continue
		}
		// 解析挂载点
		parts := strings.Fields(line)
		var mountPath string
		for i, part := range parts {
			if part == "on" && i+1 < len(parts) {
				mountPath = parts[i+1]
				break
			}
		}
		if mountPath == "" {
			continue
		}
		// 使用纯 Go 获取挂载点空间
		size, err := getDiskSize(mountPath)
		if err != nil {
			continue
		}
		if size > maxSize {
			maxSize = size
			bestMount = mountPath
		}
	}
	if bestMount == "" {
		return "", fmt.Errorf("未检测到硬盘挂载点")
	}
	return bestMount, nil
}

// detectArchitecture 检测当前系统架构
func detectArchitecture() (string, error) {
	cmd := exec.Command("uname", "-m")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("无法检测系统架构: %v", err)
	}

	arch := strings.TrimSpace(string(output))
	switch arch {
	case "aarch64", "arm64":
		return "arm64", nil
	case "x86_64", "amd64":
		return "x86_64", nil
	default:
		return "", fmt.Errorf("不支持的架构: %s", arch)
	}
}

// executeScript 执行脚本并实时输出日志
func executeScript(scriptPath string) error {
	cmd := exec.Command("sh", scriptPath)

	// 创建管道来捕获标准输出和标准错误
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建 stdout 管道失败: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建 stderr 管道失败: %v", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动脚本失败: %v", err)
	}

	// 实时打印输出
	go streamOutput(stdout)
	go streamOutput(stderr)

	// 等待命令完成
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("脚本执行失败: %v", err)
	}

	return nil
}

// streamOutput 流式输出日志
func streamOutput(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}
