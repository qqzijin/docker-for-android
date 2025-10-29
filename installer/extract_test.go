package main

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

// TestExtractTarGz 测试解压 tar.gz 文件
func TestExtractTarGz(t *testing.T) {
	// 创建测试目录
	tmpDir := t.TempDir()

	// 创建一个测试 tar.gz 文件
	tarGzPath := filepath.Join(tmpDir, "test.tar.gz")
	err := createTestTarGz(tarGzPath, map[string]string{
		"file1.txt":     "content of file1",
		"file2.txt":     "content of file2",
		"dir/file3.txt": "content of file3",
	})
	if err != nil {
		t.Fatalf("创建测试 tar.gz 失败: %v", err)
	}

	// 解压到新目录
	extractDir := filepath.Join(tmpDir, "extracted")
	err = extractTarGz(tarGzPath, extractDir, "")
	if err != nil {
		t.Fatalf("解压失败: %v", err)
	}

	// 验证文件
	tests := []struct {
		path    string
		content string
	}{
		{"file1.txt", "content of file1"},
		{"file2.txt", "content of file2"},
		{"dir/file3.txt", "content of file3"},
	}

	for _, tt := range tests {
		fullPath := filepath.Join(extractDir, tt.path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("读取文件 %s 失败: %v", tt.path, err)
			continue
		}
		if string(content) != tt.content {
			t.Errorf("文件 %s 内容不匹配，期望: %s, 实际: %s", tt.path, tt.content, string(content))
		}
	}
}

// TestExtractTarGz_EmptyArchive 测试空归档
func TestExtractTarGz_EmptyArchive(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建空 tar.gz
	tarGzPath := filepath.Join(tmpDir, "empty.tar.gz")
	err := createTestTarGz(tarGzPath, map[string]string{})
	if err != nil {
		t.Fatalf("创建空 tar.gz 失败: %v", err)
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	err = extractTarGz(tarGzPath, extractDir, "")
	if err != nil {
		t.Fatalf("解压空归档失败: %v", err)
	}
}

// TestExtractTarGz_NestedDirectories 测试嵌套目录
func TestExtractTarGz_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建包含嵌套目录的 tar.gz
	tarGzPath := filepath.Join(tmpDir, "nested.tar.gz")
	err := createTestTarGz(tarGzPath, map[string]string{
		"a/b/c/file.txt":       "nested file",
		"a/b/file2.txt":        "less nested",
		"a/file3.txt":          "even less nested",
		"x/y/z/deep/file4.txt": "very deep",
	})
	if err != nil {
		t.Fatalf("创建嵌套 tar.gz 失败: %v", err)
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	err = extractTarGz(tarGzPath, extractDir, "")
	if err != nil {
		t.Fatalf("解压失败: %v", err)
	}

	// 验证所有文件都存在
	expectedFiles := []string{
		"a/b/c/file.txt",
		"a/b/file2.txt",
		"a/file3.txt",
		"x/y/z/deep/file4.txt",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(extractDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("文件不存在: %s", file)
		}
	}
}

// TestExtractTarGz_FileNotExist 测试文件不存在
func TestExtractTarGz_FileNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	extractDir := filepath.Join(tmpDir, "extracted")

	err := extractTarGz("/nonexistent/file.tar.gz", extractDir, "")
	if err == nil {
		t.Error("期望返回错误，但没有错误")
	}
}

// TestExtractTarGz_InvalidGzip 测试无效的 gzip 文件
func TestExtractTarGz_InvalidGzip(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建一个非 gzip 文件
	invalidFile := filepath.Join(tmpDir, "invalid.tar.gz")
	err := os.WriteFile(invalidFile, []byte("not a gzip file"), 0644)
	if err != nil {
		t.Fatalf("创建无效文件失败: %v", err)
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	err = extractTarGz(invalidFile, extractDir, "")
	if err == nil {
		t.Error("期望返回错误，但没有错误")
	}
}

// TestSetBinPermissions 测试设置二进制文件权限
func TestSetBinPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := filepath.Join(tmpDir, "bin")
	err := os.MkdirAll(binDir, 0755)
	if err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}

	// 创建一些测试文件
	files := []string{"docker", "dockerd", "containerd", "runc"}
	for _, file := range files {
		path := filepath.Join(binDir, file)
		err := os.WriteFile(path, []byte("binary content"), 0644)
		if err != nil {
			t.Fatalf("创建文件 %s 失败: %v", file, err)
		}
	}

	// 设置权限
	err = setBinPermissions(binDir)
	if err != nil {
		t.Fatalf("设置权限失败: %v", err)
	}

	// 验证权限
	for _, file := range files {
		path := filepath.Join(binDir, file)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("获取文件信息失败: %v", err)
		}

		mode := info.Mode()
		expectedMode := os.FileMode(0755)

		// 在 Unix 系统上检查权限
		if mode.Perm() != expectedMode {
			t.Errorf("文件 %s 权限不正确，期望: %o, 实际: %o", file, expectedMode, mode.Perm())
		}
	}
}

// TestSetBinPermissions_WithSubdirectories 测试有子目录的情况
func TestSetBinPermissions_WithSubdirectories(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := filepath.Join(tmpDir, "bin")

	// 创建带子目录的结构
	err := os.MkdirAll(filepath.Join(binDir, "subdir"), 0755)
	if err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}

	// 创建文件
	files := map[string]string{
		"file1":        "content1",
		"file2":        "content2",
		"subdir/file3": "content3",
	}

	for file, content := range files {
		path := filepath.Join(binDir, file)
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("创建文件 %s 失败: %v", file, err)
		}
	}

	// 设置权限（只设置顶层目录的文件）
	err = setBinPermissions(binDir)
	if err != nil {
		t.Fatalf("设置权限失败: %v", err)
	}

	// 验证顶层文件权限
	for _, file := range []string{"file1", "file2"} {
		path := filepath.Join(binDir, file)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("获取文件信息失败: %v", err)
		}

		if info.Mode().Perm() != 0755 {
			t.Errorf("文件 %s 权限不正确", file)
		}
	}
}

// TestSetBinPermissions_EmptyDirectory 测试空目录
func TestSetBinPermissions_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := filepath.Join(tmpDir, "empty-bin")
	err := os.MkdirAll(binDir, 0755)
	if err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}

	err = setBinPermissions(binDir)
	if err != nil {
		t.Fatalf("设置空目录权限失败: %v", err)
	}
}

// TestSetBinPermissions_DirectoryNotExist 测试目录不存在
func TestSetBinPermissions_DirectoryNotExist(t *testing.T) {
	err := setBinPermissions("/nonexistent/directory")
	if err == nil {
		t.Error("期望返回错误，但没有错误")
	}
}

// TestExtractTarGz_PreservePermissions 测试保持文件权限
func TestExtractTarGz_PreservePermissions(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建带不同权限的文件的 tar.gz
	tarGzPath := filepath.Join(tmpDir, "perms.tar.gz")
	err := createTestTarGzWithPerms(tarGzPath, map[string]fileWithPerm{
		"executable": {content: "#!/bin/sh\necho hello", mode: 0755},
		"readonly":   {content: "read only", mode: 0444},
		"normal":     {content: "normal file", mode: 0644},
	})
	if err != nil {
		t.Fatalf("创建 tar.gz 失败: %v", err)
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	err = extractTarGz(tarGzPath, extractDir, "")
	if err != nil {
		t.Fatalf("解压失败: %v", err)
	}

	// 验证权限
	tests := []struct {
		file string
		mode os.FileMode
	}{
		{"executable", 0755},
		{"readonly", 0444},
		{"normal", 0644},
	}

	for _, tt := range tests {
		fullPath := filepath.Join(extractDir, tt.file)
		info, err := os.Stat(fullPath)
		if err != nil {
			t.Errorf("获取文件 %s 信息失败: %v", tt.file, err)
			continue
		}

		if info.Mode().Perm() != tt.mode {
			t.Errorf("文件 %s 权限不正确，期望: %o, 实际: %o", tt.file, tt.mode, info.Mode().Perm())
		}
	}
}

// BenchmarkExtractTarGz 解压性能基准测试
func BenchmarkExtractTarGz(b *testing.B) {
	tmpDir := b.TempDir()

	// 创建一个包含多个文件的 tar.gz
	files := make(map[string]string)
	for i := 0; i < 100; i++ {
		files[filepath.Join("dir", filepath.Base(filepath.Join("", string(rune('a'+i%26)))), "file.txt")] = "test content"
	}

	tarGzPath := filepath.Join(tmpDir, "bench.tar.gz")
	err := createTestTarGz(tarGzPath, files)
	if err != nil {
		b.Fatalf("创建 tar.gz 失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractDir := filepath.Join(tmpDir, "extracted", string(rune(i)))
		err := extractTarGz(tarGzPath, extractDir, "")
		if err != nil {
			b.Fatalf("解压失败: %v", err)
		}
	}
}

// Helper functions

type fileWithPerm struct {
	content string
	mode    os.FileMode
}

// createTestTarGz 创建测试用的 tar.gz 文件
func createTestTarGz(path string, files map[string]string) error {
	filesWithPerm := make(map[string]fileWithPerm)
	for name, content := range files {
		filesWithPerm[name] = fileWithPerm{
			content: content,
			mode:    0644,
		}
	}
	return createTestTarGzWithPerms(path, filesWithPerm)
}

// createTestTarGzWithPerms 创建带权限的测试 tar.gz 文件
func createTestTarGzWithPerms(path string, files map[string]fileWithPerm) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// 添加文件到归档
	for name, fileInfo := range files {
		// 如果文件在子目录中，先创建目录
		dir := filepath.Dir(name)
		if dir != "." && dir != "" {
			// 创建所有父目录
			dirs := []string{}
			for d := dir; d != "." && d != ""; d = filepath.Dir(d) {
				dirs = append([]string{d}, dirs...)
			}

			addedDirs := make(map[string]bool)
			for _, d := range dirs {
				if !addedDirs[d] {
					header := &tar.Header{
						Name:     d + "/",
						Mode:     0755,
						Typeflag: tar.TypeDir,
					}
					if err := tw.WriteHeader(header); err != nil {
						return err
					}
					addedDirs[d] = true
				}
			}
		}

		// 添加文件
		header := &tar.Header{
			Name: name,
			Mode: int64(fileInfo.mode),
			Size: int64(len(fileInfo.content)),
		}
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if _, err := tw.Write([]byte(fileInfo.content)); err != nil {
			return err
		}
	}

	return nil
}

// TestExtractTarGz_LargeFile 测试大文件解压
func TestExtractTarGz_LargeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过耗时测试")
	}

	tmpDir := t.TempDir()

	// 创建包含大文件的 tar.gz
	largeContent := make([]byte, 10*1024*1024) // 10MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	tarGzPath := filepath.Join(tmpDir, "large.tar.gz")
	err := createTestTarGz(tarGzPath, map[string]string{
		"large-file.dat": string(largeContent),
	})
	if err != nil {
		t.Fatalf("创建 tar.gz 失败: %v", err)
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	t.Logf("解压 10MB 文件...")
	err = extractTarGz(tarGzPath, extractDir, "")
	if err != nil {
		t.Fatalf("解压失败: %v", err)
	}

	// 验证文件大小
	extractedFile := filepath.Join(extractDir, "large-file.dat")
	info, err := os.Stat(extractedFile)
	if err != nil {
		t.Fatalf("获取文件信息失败: %v", err)
	}

	if info.Size() != int64(len(largeContent)) {
		t.Errorf("文件大小不匹配，期望: %d, 实际: %d", len(largeContent), info.Size())
	}
}
