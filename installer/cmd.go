package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// detectDiskMount 检测硬盘挂载点
func detectDiskMount() (string, error) {
	cmd := exec.Command("sh", "-c", "mount | grep -F '/dev/block/vold/public:259,1 on /mnt/media_rw/' | grep -F ' type ext4 ' | grep -oE '/mnt/media_rw/[^ ]+'")
	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		return "", fmt.Errorf("未检测到硬盘挂载点")
	}

	mountPoint := strings.TrimSpace(string(output))
	if mountPoint == "" {
		return "", fmt.Errorf("未检测到硬盘挂载点")
	}

	return mountPoint, nil
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
