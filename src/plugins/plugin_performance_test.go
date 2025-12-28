package main

import (
	"fmt"
	"log"
	"plugin"
	"time"
)

// 性能测试结果
type PerformanceResult struct {
	SkillName string
	TotalTime time.Duration
	AvgTime   time.Duration
	Calls     int
}

// 测试技能性能
func testSkillPerformance(skill any, params map[string]string, calls int) *PerformanceResult {
	start := time.Now()
	
	for i := 0; i < calls; i++ {
		// 调用技能
		if skillFunc, ok := skill.(func(map[string]string) (string, error)); ok {
			_, err := skillFunc(params)
			if err != nil {
				log.Printf("调用技能失败: %v", err)
			}
		}
	}
	
	totalTime := time.Since(start)
	avgTime := totalTime / time.Duration(calls)
	
	return &PerformanceResult{
		SkillName: "echo",
		TotalTime: totalTime,
		AvgTime:   avgTime,
		Calls:     calls
	}
}

// 运行性能测试
func RunPerformanceTest() error {
	fmt.Println("=== 插件性能测试 ===\n")
	
	// 加载插件
	p, err := plugin.Open("plugins/echo.so")
	if err != nil {
		return fmt.Errorf("加载插件失败: %v", err)
	}
	defer p.Close()
	
	// 获取Init函数
	initSymbol, err := p.Lookup("Init")
	if err != nil {
		return fmt.Errorf("获取Init函数失败: %v", err)
	}
	initFunc := initSymbol.(func(robot any))
	
	// 创建测试机器人
	testRobot := &TestRobot{
		skills: make(map[string]any),
	}
	
	// 初始化插件
	initFunc(testRobot)
	
	// 测试echo技能
	if skill, ok := testRobot.skills["echo"]; ok {
		params := map[string]string{"message": "test performance"}
		
		// 运行1000次测试
		result := testSkillPerformance(skill, params, 1000)
		
		fmt.Printf("技能: %s\n", result.SkillName)
		fmt.Printf("调用次数: %d\n", result.Calls)
		fmt.Printf("总耗时: %v\n", result.TotalTime)
		fmt.Printf("平均耗时: %v\n", result.AvgTime)
		fmt.Printf("每秒调用次数: %.2f\n", float64(result.Calls)/result.TotalTime.Seconds())
	}
	
	return nil
}

// TestRobot 实现了Robot接口
type TestRobot struct {
	skills map[string]any
}

func (r *TestRobot) HandleSkill(skillName string, skill any) {
	r.skills[skillName] = skill
}

func main() {
	if err := RunPerformanceTest(); err != nil {
		log.Fatalf("性能测试失败: %v", err)
	}
}