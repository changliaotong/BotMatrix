package development_team

import (
	"BotMatrix/common/ai"
	"fmt"
	"sync"
	"time"
)

type DevelopmentTeam struct {
	architect           *ArchitectImpl
	programmer          *ProgrammerImpl
	databaseSpecialist  *DatabaseSpecialistImpl
	tester              *TesterImpl
	reviewer            *ReviewerImpl
	aiSvc               ai.AIService
	lock                sync.Mutex
}

func NewDevelopmentTeam(aiSvc ai.AIService) *DevelopmentTeam {
	return &DevelopmentTeam{
		architect:           NewArchitect(aiSvc),
		programmer:          NewProgrammer(aiSvc),
		databaseSpecialist:  NewDatabaseSpecialist(aiSvc),
		tester:              NewTester(aiSvc),
		reviewer:            NewReviewer(aiSvc),
		aiSvc:               aiSvc,
	}
}

type Project struct {
	ID          string
	Name        string
	Description string
	Status      string
	Tasks       []Task
	Results     map[string]Result
}

func (dt *DevelopmentTeam) StartProject(project *Project) error {
	dt.lock.Lock()
	defer dt.lock.Unlock()

	project.Status = "in_progress"
	project.Results = make(map[string]Result)

	for _, task := range project.Tasks {
		result, err := dt.ExecuteTask(task)
		if err != nil {
			project.Status = "failed"
			return fmt.Errorf("task %s failed: %v", task.ID, err)
		}
		project.Results[task.ID] = result
	}

	project.Status = "completed"
	return nil
}

func (dt *DevelopmentTeam) ExecuteTask(task Task) (Result, error) {
	switch task.Type {
	case "design_architecture", "generate_tech_stack", "create_module_structure":
		return dt.architect.ExecuteTask(task)
	case "generate_code", "refactor_code", "fix_bug":
		return dt.programmer.ExecuteTask(task)
	case "generate_schema", "optimize_queries", "migrate_database":
		return dt.databaseSpecialist.ExecuteTask(task)
	case "generate_test_cases", "execute_tests", "generate_test_report":
		return dt.tester.ExecuteTask(task)
	case "review_code", "check_security", "enforce_best_practices":
		return dt.reviewer.ExecuteTask(task)
	default:
		return Result{}, fmt.Errorf("unknown task type: %s", task.Type)
	}
}

func (dt *DevelopmentTeam) GetTeamMembers() []DeveloperRole {
	return []DeveloperRole{
		dt.architect,
		dt.programmer,
		dt.databaseSpecialist,
		dt.tester,
		dt.reviewer,
	}
}

func (dt *DevelopmentTeam) TrainTeam(skill string, experience int) {
	for _, member := range dt.GetTeamMembers() {
		member.Learn(skill, experience)
	}
}

func (dt *DevelopmentTeam) GetTeamStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	for _, member := range dt.GetTeamMembers() {
		stats[member.GetRole()] = map[string]interface{}{
			"skills":     member.GetSkills(),
			"experience": member.GetExperience(),
		}
	}
	
	return stats
}

func (dt *DevelopmentTeam) AutoEvolve() {
	// 这里可以实现团队的自动进化逻辑
	// 例如根据历史任务结果优化团队成员的技能
	fmt.Println("Team is evolving automatically...")
}