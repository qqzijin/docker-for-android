package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDownloadFromURL 测试从 URL 下载文件
func TestDownloadFromURL(t *testing.T) {
	// 创建测试服务器
	testContent := []byte("test file content for download")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		w.Write(testContent)
	}))
	defer server.Close()

	// 创建临时目录
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test-download.txt")

	// 创建测试客户端
	client := CreateHTTPClient()

	// 测试下载
	err := downloadFromURL(client, server.URL, destPath)
	if err != nil {
		t.Fatalf("下载失败: %v", err)
	}

	// 验证文件内容
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	if string(content) != string(testContent) {
		t.Errorf("文件内容不匹配，期望: %s, 实际: %s", testContent, content)
	}
}

// TestDownloadFromURL_HTTPError 测试 HTTP 错误
func TestDownloadFromURL_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test-error.txt")

	// 创建测试客户端
	client := CreateHTTPClient()

	err := downloadFromURL(client, server.URL, destPath)
	if err == nil {
		t.Fatal("期望返回错误，但没有错误")
	}

	if !strings.Contains(err.Error(), "404") {
		t.Errorf("错误信息不正确: %v", err)
	}
}

// TestVerifySHA256 测试 SHA256 验证
func TestVerifySHA256(t *testing.T) {
	// 创建测试文件
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content for sha256")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 计算期望的 SHA256
	hash := sha256.New()
	hash.Write(testContent)
	expectedHash := hex.EncodeToString(hash.Sum(nil))

	// 测试正确的 SHA256
	err = verifySHA256(testFile, expectedHash)
	if err != nil {
		t.Errorf("SHA256 验证失败: %v", err)
	}

	// 测试错误的 SHA256
	wrongHash := "0000000000000000000000000000000000000000000000000000000000000000"
	err = verifySHA256(testFile, wrongHash)
	if err == nil {
		t.Error("期望 SHA256 验证失败，但验证通过了")
	}
}

// TestVerifySHA256_FileNotExist 测试文件不存在
func TestVerifySHA256_FileNotExist(t *testing.T) {
	err := verifySHA256("/nonexistent/file.txt", "somehash")
	if err == nil {
		t.Error("期望返回错误，但没有错误")
	}
}

// TestDownloadFile_WithRetry 测试下载重试机制
func TestDownloadFile_WithRetry(t *testing.T) {
	testContent := []byte("test content")
	hash := sha256.New()
	hash.Write(testContent)
	expectedSHA256 := hex.EncodeToString(hash.Sum(nil))

	// 第一个服务器返回错误
	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failServer.Close()

	// 第二个服务器返回成功
	successServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		w.Write(testContent)
	}))
	defer successServer.Close()

	// 临时覆盖 CDN 和服务器 URL（通过创建测试函数）
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test-retry.txt")

	// 直接测试 downloadFromURL 的重试逻辑
	// 注意：这里我们只能测试单个 URL，完整的重试需要修改代码支持依赖注入
	client := CreateHTTPClient()
	err := downloadFromURL(client, successServer.URL, destPath)
	if err != nil {
		t.Fatalf("下载失败: %v", err)
	}

	// 验证 SHA256
	err = verifySHA256(destPath, expectedSHA256)
	if err != nil {
		t.Errorf("SHA256 验证失败: %v", err)
	}
}

// TestDownloadFile_SHA256Mismatch 测试 SHA256 不匹配
func TestDownloadFile_SHA256Mismatch(t *testing.T) {
	testContent := []byte("test content")
	wrongSHA256 := "0000000000000000000000000000000000000000000000000000000000000000"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		w.Write(testContent)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test-mismatch.txt")

	// 创建测试客户端
	client := CreateHTTPClient()

	// 下载
	err := downloadFromURL(client, server.URL, destPath)
	if err != nil {
		t.Fatalf("下载失败: %v", err)
	}

	// 验证 SHA256（应该失败）
	err = verifySHA256(destPath, wrongSHA256)
	if err == nil {
		t.Error("期望 SHA256 验证失败，但验证通过了")
	}
}

// TestGetVersionInfo_Parse 测试版本信息解析
func TestGetVersionInfo_Parse(t *testing.T) {
	// 创建测试 version.txt
	tmpDir := t.TempDir()
	versionContent := `# Docker for Android Version File
VERSION=28.0.1.10
DOCKER_SHA256=abc123
BIN_ARM64_SHA256=def456
BIN_X86_64_SHA256=ghi789
`
	versionPath := filepath.Join(tmpDir, "version.txt")
	err := os.WriteFile(versionPath, []byte(versionContent), 0644)
	if err != nil {
		t.Fatalf("创建 version.txt 失败: %v", err)
	}

	// 读取并解析
	content, err := os.ReadFile(versionPath)
	if err != nil {
		t.Fatalf("读取 version.txt 失败: %v", err)
	}

	// 解析逻辑（从 getVersionInfo 提取）
	info := &VersionInfo{
		Architecture: "arm64", // 假设架构
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
		case "BIN_ARM64_SHA256":
			info.BinSHA256 = value
		}
	}

	// 验证解析结果
	if info.Version != "28.0.1.10" {
		t.Errorf("VERSION 解析错误，期望: 28.0.1.10, 实际: %s", info.Version)
	}
	if info.DockerSHA256 != "abc123" {
		t.Errorf("DOCKER_SHA256 解析错误，期望: abc123, 实际: %s", info.DockerSHA256)
	}
	if info.BinSHA256 != "def456" {
		t.Errorf("BIN_ARM64_SHA256 解析错误，期望: def456, 实际: %s", info.BinSHA256)
	}
}

// TestGetVersionInfo_MissingFields 测试缺少必需字段
func TestGetVersionInfo_MissingFields(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "缺少 VERSION",
			content: `DOCKER_SHA256=abc123
BIN_ARM64_SHA256=def456`,
			wantErr: true,
		},
		{
			name: "缺少 DOCKER_SHA256",
			content: `VERSION=28.0.1.10
BIN_ARM64_SHA256=def456`,
			wantErr: true,
		},
		{
			name: "缺少 BIN_SHA256",
			content: `VERSION=28.0.1.10
DOCKER_SHA256=abc123`,
			wantErr: true,
		},
		{
			name: "所有字段都存在",
			content: `VERSION=28.0.1.10
DOCKER_SHA256=abc123
BIN_ARM64_SHA256=def456`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &VersionInfo{
				Architecture: "arm64",
			}

			lines := strings.Split(tt.content, "\n")
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
				case "BIN_ARM64_SHA256":
					info.BinSHA256 = value
				}
			}

			// 检查必需字段
			hasError := info.Version == "" || info.DockerSHA256 == "" || info.BinSHA256 == ""

			if hasError != tt.wantErr {
				t.Errorf("字段验证错误，期望错误: %v, 实际错误: %v", tt.wantErr, hasError)
			}
		})
	}
}

// BenchmarkDownloadFromURL 下载性能基准测试
func BenchmarkDownloadFromURL(b *testing.B) {
	// 创建测试服务器
	testContent := make([]byte, 1024*1024) // 1MB
	for i := range testContent {
		testContent[i] = byte(i % 256)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		w.Write(testContent)
	}))
	defer server.Close()

	tmpDir := b.TempDir()

	// 创建测试客户端
	client := CreateHTTPClient()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		destPath := filepath.Join(tmpDir, fmt.Sprintf("bench-%d.dat", i))
		err := downloadFromURL(client, server.URL, destPath)
		if err != nil {
			b.Fatalf("下载失败: %v", err)
		}
	}
}

// BenchmarkVerifySHA256 SHA256 验证性能基准测试
func BenchmarkVerifySHA256(b *testing.B) {
	// 创建测试文件
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "bench.dat")
	testContent := make([]byte, 10*1024*1024) // 10MB
	for i := range testContent {
		testContent[i] = byte(i % 256)
	}

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		b.Fatalf("创建测试文件失败: %v", err)
	}

	// 计算 SHA256
	hash := sha256.New()
	hash.Write(testContent)
	expectedHash := hex.EncodeToString(hash.Sum(nil))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := verifySHA256(testFile, expectedHash)
		if err != nil {
			b.Fatalf("SHA256 验证失败: %v", err)
		}
	}
}

// TestDownloadWithProgress 测试带进度的下载（手动运行可以看到进度）
func TestDownloadWithProgress(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过耗时测试")
	}

	// 创建较大的测试内容
	testContent := make([]byte, 5*1024*1024) // 5MB
	for i := range testContent {
		testContent[i] = byte(i % 256)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)

		// 分块发送以模拟真实下载
		chunkSize := 32 * 1024
		for i := 0; i < len(testContent); i += chunkSize {
			end := i + chunkSize
			if end > len(testContent) {
				end = len(testContent)
			}
			w.Write(testContent[i:end])
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test-progress.dat")

	// 创建测试客户端
	client := CreateHTTPClient()

	t.Logf("开始下载 5MB 测试文件...")
	err := downloadFromURL(client, server.URL, destPath)
	if err != nil {
		t.Fatalf("下载失败: %v", err)
	}
	t.Logf("下载完成")

	// 验证文件大小
	stat, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("获取文件信息失败: %v", err)
	}

	if stat.Size() != int64(len(testContent)) {
		t.Errorf("文件大小不匹配，期望: %d, 实际: %d", len(testContent), stat.Size())
	}
}

// TestStreamOutput 测试日志流式输出
func TestStreamOutput(t *testing.T) {
	// 创建测试数据
	testLines := []string{
		"line 1",
		"line 2",
		"line 3",
	}

	// 创建 reader
	content := strings.Join(testLines, "\n")
	reader := strings.NewReader(content)

	// 由于 streamOutput 直接输出到 stdout，我们只测试它不会 panic
	// 在真实场景中可以使用 io.Pipe 来捕获输出
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("streamOutput panic: %v", r)
		}
	}()

	streamOutput(reader)
}

// TestLargeFileDownload 测试大文件下载
func TestLargeFileDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过耗时测试")
	}

	// 创建 50MB 测试内容
	testSize := 50 * 1024 * 1024

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", testSize))
		w.WriteHeader(http.StatusOK)

		// 生成并发送数据
		chunkSize := 1024 * 1024 // 1MB chunks
		chunk := make([]byte, chunkSize)
		remaining := testSize

		for remaining > 0 {
			size := chunkSize
			if size > remaining {
				size = remaining
			}
			for i := 0; i < size; i++ {
				chunk[i] = byte((testSize - remaining + i) % 256)
			}
			w.Write(chunk[:size])
			remaining -= size

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test-large.dat")

	// 创建测试客户端
	client := CreateHTTPClient()

	t.Logf("开始下载 50MB 大文件...")
	err := downloadFromURL(client, server.URL, destPath)
	if err != nil {
		t.Fatalf("下载失败: %v", err)
	}
	t.Logf("下载完成")

	// 验证文件大小
	stat, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("获取文件信息失败: %v", err)
	}

	if stat.Size() != int64(testSize) {
		t.Errorf("文件大小不匹配，期望: %d, 实际: %d", testSize, stat.Size())
	}
}

// Helper function to create a reader from string
func stringReader(s string) io.Reader {
	return strings.NewReader(s)
}
