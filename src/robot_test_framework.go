package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// TestResult 测试结果
type TestResult struct {
	PluginName string        `json:"plugin_name"`
	TestName   string        `json:"test_name"`
	Success    bool          `json:"success"`
	Error      string        `json:"error,omitempty"`
	Output     string        `json:"output,omitempty"`
	Duration   time.Duration `json:"duration"`
}

// TestCase 测试用例
type TestCase struct {
	Name     string `json:"name"`
	Input    string `json:"input"`
	Expected string `json:"expected"`
	Timeout  int    `json:"timeout"`
}

// TestFramework 测试框架
type TestFramework struct {
	PluginDir  string     `json:"plugin_dir"`
	ManualMode bool       `json:"manual_mode"`
	TestCases  []TestCase `json:"test_cases"`
}

// NewTestFramework 创建测试框架
func NewTestFramework(pluginDir string) *TestFramework {
	return &TestFramework{
		PluginDir:  pluginDir,
		ManualMode: false,
		TestCases:  []TestCase{},
	}
}

// LoadTestCases 加载测试用例
func (tf *TestFramework) LoadTestCases(testFile string) error {
	file, err := os.Open(testFile)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&tf.TestCases); err != nil {
		return err
	}

	return nil
}

// RunPluginTest 运行插件测试
func (tf *TestFramework) RunPluginTest(pluginName string, testCase TestCase) TestResult {
	startTime := time.Now()

	// 构建插件路径
	pluginPath := filepath.Join(tf.PluginDir, pluginName, fmt.Sprintf("%s.exe", pluginName))
	// 检查是否存在插件可执行文件
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		// 尝试直接使用插件名称（假设已在PATH中）
		pluginPath = pluginName
	} else {
		// 使用绝对路径
		pluginPath, _ = filepath.Abs(pluginPath)
	}

	// 检查插件是否存在（仅当不是直接使用插件名称时）
	if pluginPath != pluginName {
		if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
			return TestResult{
				PluginName: pluginName,
				TestName:   testCase.Name,
				Success:    false,
				Error:      fmt.Sprintf("Plugin not found: %s", pluginPath),
				Duration:   time.Since(startTime),
			}
		}
	}

	// 运行插件测试
	cmd := exec.Command(pluginPath, testCase.Input)
	// 设置工作目录
	cmd.Dir = filepath.Join(tf.PluginDir, pluginName)
	// 捕获标准输出和错误
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// 运行命令
	err := cmd.Run()
	outputStr := stdout.String() + stderr.String()
	// 调试输出
	fmt.Printf("Plugin: %s, Input: %s, Output: %q, Error: %v\n", pluginName, testCase.Input, outputStr, err)

	// 检查测试结果
	success := err == nil
	if success && testCase.Expected != "" {
		success = strings.Contains(outputStr, testCase.Expected)
		if !success {
			err = fmt.Errorf("Expected '%s' not found in output", testCase.Expected)
		}
	}

	return TestResult{
		PluginName: pluginName,
		TestName:   testCase.Name,
		Success:    success,
		Error:      fmt.Sprintf("%v", err),
		Output:     outputStr,
		Duration:   time.Since(startTime),
	}
}

// RunManualTest 运行手动测试
func (tf *TestFramework) RunManualTest(pluginName string) {
	fmt.Printf("=== Manual Test for Plugin: %s ===\n", pluginName)
	fmt.Println("Enter commands to test the plugin (type 'exit' to quit)")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		if strings.ToLower(input) == "exit" {
			break
		}

		// 运行插件测试
		result := tf.RunPluginTest(pluginName, TestCase{
			Name:     "manual_test",
			Input:    input,
			Expected: "",
			Timeout:  5000,
		})

		fmt.Printf("\nResult: %s\n", map[bool]string{true: "✓ Success", false: "✗ Failed"}[result.Success])
		fmt.Printf("Output: %s\n", result.Output)
		if !result.Success {
			fmt.Printf("Error: %s\n", result.Error)
		}
	}
}

// RunAutomatedTests 运行自动化测试
func (tf *TestFramework) RunAutomatedTests(pluginName string) []TestResult {
	fmt.Printf("=== Automated Tests for Plugin: %s ===\n", pluginName)

	var results []TestResult
	for _, testCase := range tf.TestCases {
		fmt.Printf("\nRunning test: %s...\n", testCase.Name)
		result := tf.RunPluginTest(pluginName, testCase)
		results = append(results, result)

		status := map[bool]string{true: "✓ Success", false: "✗ Failed"}[result.Success]
		fmt.Printf("%s (%.2fs)\n", status, result.Duration.Seconds())
		if !result.Success {
			fmt.Printf("Error: %s\n", result.Error)
		}
	}

	// 统计结果
	passCount := 0
	for _, result := range results {
		if result.Success {
			passCount++
		}
	}

	fmt.Printf("\n=== Test Summary ===\n")
	fmt.Printf("Total tests: %d\n", len(results))
	fmt.Printf("Passed: %d\n", passCount)
	fmt.Printf("Failed: %d\n", len(results)-passCount)
	fmt.Printf("Pass rate: %.1f%%\n", float64(passCount)/float64(len(results))*100)

	return results
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run robot_test_framework.go <plugin_name> [--manual | --test-file <test_file>]")
		os.Exit(1)
	}

	pluginName := os.Args[1]
	framework := NewTestFramework("src/plugins")

	// 处理命令行参数
	if len(os.Args) > 2 {
		if os.Args[2] == "--manual" {
			framework.RunManualTest(pluginName)
			return
		} else if os.Args[2] == "--test-file" && len(os.Args) > 3 {
			if err := framework.LoadTestCases(os.Args[3]); err != nil {
				log.Fatalf("Failed to load test cases: %v", err)
			}
			framework.RunAutomatedTests(pluginName)
			return
		}
	}

	// 默认运行手动测试
	framework.RunManualTest(pluginName)
}
