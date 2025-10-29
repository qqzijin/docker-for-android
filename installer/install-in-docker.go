package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// 安装路径
	dockerRoot = "/data/local/docker"
	binDir     = "/data/local/docker/bin"

	// 版本文件
	versionFile = "version.txt"
)

type VersionInfo struct {
	Version      string
	DockerSHA256 string
	BinSHA256    string
	Architecture string
}

func main() {
	fmt.Println("==========================================")
	fmt.Println("Docker for Android - Installer")
	fmt.Println("==========================================")
	fmt.Println()

	// Step 1: 检测硬盘挂载
	fmt.Println("[1/5] 检测硬盘挂载点...")
	diskRoot, err := detectDiskMount()
	if err != nil {
		fmt.Printf("✗ 错误: %v\n", err)
		fmt.Println("✗ Docker 需要 ext4 格式的外置硬盘才能运行")
		fmt.Println("✗ 请确保已接入并格式化硬盘后再运行此程序")
		os.Exit(1)
	}
	fmt.Printf("✓ 检测到硬盘挂载点: %s\n", diskRoot)
	fmt.Println()

	// 设置临时文件夹
	tmpDir := filepath.Join(diskRoot, "Cache", "installer")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		fmt.Printf("✗ 错误: 无法创建临时目录 %s: %v\n", tmpDir, err)
		os.Exit(1)
	}
	fmt.Printf("✓ 临时目录: %s\n", tmpDir)
	fmt.Println()

	// Step 2: 获取版本信息
	fmt.Println("[2/5] 获取版本信息...")
	version, err := getVersionInfo(tmpDir)
	if err != nil {
		fmt.Printf("✗ 错误: 无法获取版本信息: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ 版本: %s\n", version.Version)
	fmt.Printf("✓ 架构: %s\n", version.Architecture)
	fmt.Println()

	// Step 3: 下载文件
	fmt.Println("[3/5] 下载安装文件...")

	// 下载 docker 通用包
	dockerTarFile := fmt.Sprintf("docker-%s.tar.gz", version.Version)
	dockerTarPath := filepath.Join(tmpDir, dockerTarFile)
	fmt.Printf("⏳ 下载 %s...\n", dockerTarFile)
	if err := downloadFile(dockerTarPath, dockerTarFile, version.DockerSHA256); err != nil {
		fmt.Printf("✗ 错误: 下载失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ %s 下载完成\n", dockerTarFile)

	// 下载架构特定二进制包
	binTarFile := fmt.Sprintf("docker-for-android-bin-%s-%s.tar.gz", version.Version, version.Architecture)
	binTarPath := filepath.Join(tmpDir, binTarFile)
	fmt.Printf("⏳ 下载 %s...\n", binTarFile)
	if err := downloadFile(binTarPath, binTarFile, version.BinSHA256); err != nil {
		fmt.Printf("✗ 错误: 下载失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ %s 下载完成\n", binTarFile)
	fmt.Println()

	// Step 4: 解压文件
	fmt.Println("[4/5] 解压安装文件...")

	// 解压 docker 包到 /data/local/docker
	fmt.Printf("⏳ 解压 %s 到 %s...\n", dockerTarFile, dockerRoot)
	if err := extractTarGz(dockerTarPath, "/data/local"); err != nil {
		fmt.Printf("✗ 错误: 解压失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ %s 解压完成\n", dockerTarFile)

	// 解压二进制包到 /data/local/docker/bin
	fmt.Printf("⏳ 解压 %s 到 %s...\n", binTarFile, binDir)
	if err := extractTarGz(binTarPath, binDir); err != nil {
		fmt.Printf("✗ 错误: 解压失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ %s 解压完成\n", binTarFile)

	// 设置二进制文件权限
	if err := setBinPermissions(binDir); err != nil {
		fmt.Printf("⚠ 警告: 设置二进制文件权限失败: %v\n", err)
	}
	fmt.Println()

	// Step 5: 执行部署脚本
	fmt.Println("[5/5] 执行部署脚本...")
	deployScript := filepath.Join(dockerRoot, "deploy-in-android.sh")
	if _, err := os.Stat(deployScript); os.IsNotExist(err) {
		fmt.Printf("✗ 错误: 部署脚本不存在: %s\n", deployScript)
		os.Exit(1)
	}

	if err := executeScript(deployScript); err != nil {
		fmt.Printf("✗ 错误: 部署脚本执行失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()

	// 清理临时文件
	fmt.Println("⏳ 清理临时文件...")
	os.RemoveAll(tmpDir)
	fmt.Println("✓ 清理完成")
	fmt.Println()

	fmt.Println("==========================================")
	fmt.Println("安装完成！")
	fmt.Println("==========================================")
}

// getVersionInfo 获取版本信息
func getVersionInfo(tmpDir string) (*VersionInfo, error) {
	// 检测当前架构
	arch, err := detectArchitecture()
	if err != nil {
		return nil, err
	}

	// 先尝试从 CDN 下载 version.txt
	versionPath := filepath.Join(tmpDir, versionFile)
	if err := downloadFile(versionPath, versionFile, ""); err != nil {
		return nil, fmt.Errorf("无法下载 version.txt: %v", err)
	}

	// 解析 version.txt
	content, err := os.ReadFile(versionPath)
	if err != nil {
		return nil, fmt.Errorf("无法读取 version.txt: %v", err)
	}

	info := &VersionInfo{
		Architecture: arch,
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "VERSION":
			info.Version = value
		case "DOCKER_SHA256":
			info.DockerSHA256 = value
		case fmt.Sprintf("BIN_%s_SHA256", strings.ToUpper(arch)):
			info.BinSHA256 = value
		}
	}

	if info.Version == "" {
		return nil, fmt.Errorf("version.txt 中缺少 VERSION 字段")
	}
	if info.DockerSHA256 == "" {
		return nil, fmt.Errorf("version.txt 中缺少 DOCKER_SHA256 字段")
	}
	if info.BinSHA256 == "" {
		return nil, fmt.Errorf("version.txt 中缺少 BIN_%s_SHA256 字段", strings.ToUpper(arch))
	}

	return info, nil
}
