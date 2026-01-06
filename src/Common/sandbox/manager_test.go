package sandbox

import (
	"context"
	"strings"
	"testing"
	"time"

	clog "BotMatrix/common/log"
	"BotMatrix/common/utils"

	"github.com/stretchr/testify/assert"
)

// TestSandboxIntegration 测试沙盒集成的核心功能
// 前提：本地必须运行 Docker
func TestSandboxIntegration(t *testing.T) {
	// 1. 初始化日志
	clog.InitDefaultLogger()

	// 2. 初始化 Docker 客户端
	dockerCli, err := utils.InitDockerClient()
	if err != nil {
		t.Skipf("Docker client not available, skipping integration test: %v", err)
	}

	// 2. 创建 SandboxManager
	// 使用 python:3.10-slim 镜像，如果本地没有会自动拉取
	manager := NewSandboxManager(dockerCli, "python:3.10-slim")
	ctx := context.Background()

	// 3. 创建沙盒
	t.Log("Creating sandbox...")
	sandbox, err := manager.CreateSandbox(ctx, "")
	if err != nil {
		t.Fatalf("Failed to create sandbox: %v", err)
	}
	defer func() {
		t.Log("Destroying sandbox...")
		if err := sandbox.Destroy(ctx); err != nil {
			t.Errorf("Failed to destroy sandbox: %v", err)
		}
	}()

	t.Logf("Sandbox created with ID: %s", sandbox.ID)

	// 4. 测试命令执行 (Exec)
	t.Run("Exec Command", func(t *testing.T) {
		stdout, stderr, err := sandbox.Exec(ctx, "echo 'Hello, Sandbox!'")
		assert.NoError(t, err)
		assert.Contains(t, stdout, "Hello, Sandbox!")
		assert.Empty(t, stderr)
	})

	// 5. 测试文件写入与 Python 执行
	t.Run("Write File and Run Python", func(t *testing.T) {
		code := `
print("Calculation Result:", 123 * 456)
with open("output.txt", "w") as f:
    f.write("File Created Successfully")
`
		// 写入 Python 脚本
		err := sandbox.WriteFile(ctx, "/app/script.py", []byte(code))
		assert.NoError(t, err)

		// 执行脚本
		stdout, stderr, err := sandbox.Exec(ctx, "python /app/script.py")
		assert.NoError(t, err)
		assert.Contains(t, stdout, "Calculation Result: 56088")
		if stderr != "" {
			t.Logf("Python stderr: %s", stderr)
		}

		// 验证文件是否生成
		content, err := sandbox.ReadFile(ctx, "/app/output.txt")
		assert.NoError(t, err)
		assert.Equal(t, "File Created Successfully", string(content))
	})

	// 6. 测试资源限制 (可选，稍微复杂一点)
	t.Run("Resource Limit Check", func(t *testing.T) {
		// 检查内存限制是否生效 (默认 512MB)
		// 在容器内读取 cgroup 信息
		stdout, _, err := sandbox.Exec(ctx, "cat /sys/fs/cgroup/memory.max")
		if err == nil && strings.TrimSpace(stdout) != "max" {
			// 如果能读到，验证数值。注意：不同 Docker 版本路径可能不同
			// 这里只是简单尝试
			t.Logf("Memory Limit: %s", stdout)
		}
	})

	// 暂停一小会儿以便观察 (如果是在本地调试)
	time.Sleep(1 * time.Second)
}
