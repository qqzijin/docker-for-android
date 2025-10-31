package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
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

// getFreeSpace 获取指定路径的可用空间（KB）
func getFreeSpace(path string) (uint64, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return 0, err
	}
	// 计算可用空间（字节），然后转换为 KB
	// Bfree * BlockSize / 1024
	freeSpaceKB := (stat.Bfree * uint64(stat.Bsize)) / 1024
	return freeSpaceKB, nil
}

// detectDiskMount 检测可用的存储空间（支持外接硬盘和本地空间）
func detectDiskMount() (string, error) {
	// 定义候选的存储路径（按优先级排序）
	candidatePaths := []string{
		// 外接硬盘路径
		"/mnt/media_rw",           // 外接硬盘挂载点
		"/storage",                // 存储设备
		"/mnt/sdcard",             // SD卡
		// 本地空间路径
		"/data/local/tmp",         // 系统临时目录
		"/data/data",              // 应用数据目录
		"/data",                   // 系统数据分区
		"/sdcard",                 // 内部存储
		"/storage/emulated/0",     // 模拟存储
	}

	var maxFreeSpace uint64
	var bestPath string

	// 检查每个候选路径的可用空间
	for _, basePath := range candidatePaths {
		// 检查路径是否存在
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			continue
		}

		// 获取可用空间
		freeSpace, err := getFreeSpace(basePath)
		if err != nil {
			continue
		}

		// 设置最小空间要求（1GB = 1048576 KB）
		minRequiredSpace := uint64(1048576)
		if freeSpace < minRequiredSpace {
			continue
		}

		fmt.Printf("  检测路径: %s, 可用空间: %.2f GB\n", 
			basePath, float64(freeSpace)/1024/1024)

		if freeSpace > maxFreeSpace {
			maxFreeSpace = freeSpace
			bestPath = basePath
		}
	}

	if bestPath == "" {
		// 如果没有找到合适的路径，尝试使用当前工作目录
		currentDir, err := os.Getwd()
		if err == nil {
			if freeSpace, err := getFreeSpace(currentDir); err == nil && freeSpace > 1048576 {
				fmt.Printf("  使用当前目录: %s, 可用空间: %.2f GB\n", 
					currentDir, float64(freeSpace)/1024/1024)
				return currentDir, nil
			}
		}
		return "", fmt.Errorf("未找到足够的可用空间（需要至少 1GB 空间）")
	}

	fmt.Printf("✓ 选择存储路径: %s, 可用空间: %.2f GB\n", 
		bestPath, float64(maxFreeSpace)/1024/1024)
	return bestPath, nil
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

// executeScript 执行脚本并实时输出日志，同时传递 diskRoot 参数
func executeScript(scriptPath string, diskRoot string) error {
	cmd := exec.Command("sh", scriptPath, diskRoot)

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

// stopSupervisord 停止正在运行的 supervisord 服务
// 在解压文件之前调用，避免后续服务启动失败
func stopSupervisord() error {
	supervisordPath := "/data/local/docker/bin/supervisord"

	// 检查 supervisord 文件是否存在
	if _, err := os.Stat(supervisordPath); os.IsNotExist(err) {
		// 文件不存在，说明是首次安装，无需停止服务
		fmt.Println("✓ 未检测到已安装的 supervisord，跳过停止服务")
		return nil
	}

	fmt.Println("⏳ 检测到已安装的 supervisord，正在停止服务...")

	// 执行 supervisorctl stop all
	cmd := exec.Command(supervisordPath, "ctl", "stop", "all")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 如果命令执行失败，可能是 supervisord 未运行，继续执行
		fmt.Printf("⚠ supervisorctl stop all 执行警告: %v\n", err)
		fmt.Printf("  输出: %s\n", strings.TrimSpace(string(output)))
	} else {
		fmt.Println("✓ supervisord 服务已停止")
		fmt.Printf("  %s\n", strings.TrimSpace(string(output)))
	}

	// 等待一小段时间，让进程有机会正常退出
	time.Sleep(2 * time.Second)

	// 检查 supervisord 进程是否仍在运行，并强制终止
	fmt.Println("⏳ 检查并终止 supervisord 进程...")
	if err := killSupervisordProcesses(); err != nil {
		fmt.Printf("终止 supervisord 进程失败: %v，但继续运行\n", err)
	}

	fmt.Println("✓ 所有 supervisord 进程已终止")
	return nil
}

// killSupervisordProcesses 使用 pkill 终止所有 supervisord 进程
func killSupervisordProcesses() error {
	// 使用 pkill -9 直接终止所有 supervisord 进程
	// pkill 返回码: 0=成功终止进程, 1=没有匹配的进程, 其他=错误
	cmd := exec.Command("pkill", "-9", "supervisord")
	err := cmd.Run()

	if err != nil {
		// 检查是否是 exit status 1（没有匹配的进程）
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				// 没有找到进程，这也是成功的情况
				fmt.Println("  未检测到 supervisord 进程")
				return nil
			}
		}
		// 其他错误
		return fmt.Errorf("执行 pkill 失败: %v", err)
	}

	// 等待进程完全终止
	time.Sleep(1 * time.Second)

	fmt.Println("  已终止所有 supervisord 进程")
	return nil
}
