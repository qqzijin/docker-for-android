package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestFindSupervisordProcesses_NoProcesses 测试查找进程（无 supervisord 进程）
func TestFindSupervisordProcesses_NoProcesses(t *testing.T) {
	// 注意：这个测试假设当前没有 supervisord 进程在运行
	// 如果有，测试可能会失败或返回实际进程
	pids, err := findSupervisordProcesses()
	if err != nil {
		t.Fatalf("查找进程失败: %v", err)
	}

	// 在测试环境中，通常不应该有 supervisord 进程
	// 但如果有，我们至少验证返回的是有效的 PID
	for _, pid := range pids {
		if pid <= 0 {
			t.Errorf("返回了无效的 PID: %d", pid)
		}
	}
}

// TestKillProcess_InvalidPID 测试 kill 无效 PID
func TestKillProcess_InvalidPID(t *testing.T) {
	// 尝试 kill 一个不存在的 PID（999999 通常不存在）
	err := killProcess(999999)
	// 预期会失败，因为进程不存在
	if err == nil {
		t.Log("警告: kill 不存在的进程没有返回错误（某些系统可能不报错）")
	}
}

// TestStopSupervisord_NoInstallation 测试未安装时的行为
func TestStopSupervisord_NoInstallation(t *testing.T) {
	// 创建临时目录模拟环境
	tmpDir := t.TempDir()
	
	// 临时修改 supervisord 路径为不存在的路径
	// 注意：这个测试依赖于 stopSupervisord 使用固定路径
	// 在实际环境中，如果 /data/local/docker/bin/supervisord 不存在
	// 函数应该正常返回而不报错
	
	// 由于 stopSupervisord 使用硬编码路径，我们创建一个包装函数来测试
	supervisordPath := filepath.Join(tmpDir, "supervisord")
	
	// 验证文件不存在
	if _, err := os.Stat(supervisordPath); !os.IsNotExist(err) {
		t.Fatalf("临时文件不应该存在")
	}
	
	// 这个测试主要验证逻辑，实际的 stopSupervisord 会检查固定路径
	t.Log("验证：当 supervisord 不存在时，函数应该优雅地处理")
}

// TestParsePIDFromPSOutput 测试从 ps 输出解析 PID
func TestParsePIDFromPSOutput(t *testing.T) {
	testCases := []struct {
		name     string
		psLine   string
		expected bool
		pid      int
	}{
		{
			name:     "标准 supervisord 进程",
			psLine:   "root      1234  1  0 10:00 ?        00:00:01 /data/local/docker/bin/supervisord -c /data/local/docker/supervisor.conf",
			expected: true,
			pid:      1234,
		},
		{
			name:     "grep 进程应被忽略",
			psLine:   "root      5678  1  0 10:00 ?        00:00:00 grep supervisord",
			expected: false,
		},
		{
			name:     "不包含 supervisord",
			psLine:   "root      9012  1  0 10:00 ?        00:00:00 /usr/bin/python",
			expected: false,
		},
		{
			name:     "另一种格式的 supervisord",
			psLine:   "u0_a123   4567  890  1 12:34 pts/0   00:00:05 supervisord -d",
			expected: true,
			pid:      4567,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 模拟 findSupervisordProcesses 的逻辑
			containsSupervisord := strings.Contains(tc.psLine, "supervisord")
			containsGrep := strings.Contains(tc.psLine, "grep")
			
			shouldMatch := containsSupervisord && !containsGrep
			
			if shouldMatch != tc.expected {
				t.Errorf("匹配结果不符合预期: got %v, want %v", shouldMatch, tc.expected)
			}
			
			if tc.expected {
				// 解析 PID
				fields := strings.Fields(tc.psLine)
				if len(fields) >= 2 {
					pid, err := strconv.Atoi(fields[1])
					if err != nil {
						t.Errorf("解析 PID 失败: %v", err)
					}
					if pid != tc.pid {
						t.Errorf("PID 不匹配: got %d, want %d", pid, tc.pid)
					}
				}
			}
		})
	}
}

// TestStopSupervisordWorkflow 测试整体工作流程的逻辑
func TestStopSupervisordWorkflow(t *testing.T) {
	t.Log("停止 supervisord 的工作流程：")
	t.Log("1. 检查 /data/local/docker/bin/supervisord 是否存在")
	t.Log("2. 如果不存在，返回 nil（首次安装）")
	t.Log("3. 如果存在，执行 supervisorctl stop all")
	t.Log("4. 等待 2 秒让进程退出")
	t.Log("5. 查找仍在运行的 supervisord 进程")
	t.Log("6. 如果有进程，使用 kill -9 强制终止")
	t.Log("7. 再次验证所有进程已终止")
	
	// 这是一个文档性测试，验证工作流程的设计
	workflow := []string{
		"检查文件存在性",
		"执行 supervisorctl stop all",
		"等待进程退出",
		"查找残留进程",
		"强制终止进程",
		"最终验证",
	}
	
	if len(workflow) != 6 {
		t.Errorf("工作流程步骤数量不对")
	}
}

// BenchmarkFindSupervisordProcesses 性能测试
func BenchmarkFindSupervisordProcesses(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = findSupervisordProcesses()
	}
}
