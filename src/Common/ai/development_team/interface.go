package development_team

type DeveloperRole interface {
    GetRole() string
    ExecuteTask(task Task) (Result, error)
    GetSkills() []string
    GetExperience() int
    Learn(skill string, experience int)
}

type Task struct {
    ID          string
    Type        string
    Description string
    Input       map[string]interface{}
    Priority    int
}

type Result struct {
    Success     bool
    Output      map[string]interface{}
    Log         string
    Error       error
    ExecutionTime float64
}

type Architect interface {
    DeveloperRole
    DesignArchitecture(requirements string) string
    GenerateTechStack() []string
    CreateModuleStructure() map[string]interface{}
}

type Programmer interface {
    DeveloperRole
    GenerateCode(prompt string, language string) string
    RefactorCode(code string, improvements []string) string
    FixBug(code string, bugDescription string) string
}

type DatabaseSpecialist interface {
    DeveloperRole
    GenerateDatabaseSchema(requirements string) string
    OptimizeQueries(queries []string) []string
    MigrateDatabase(oldSchema string, newSchema string) string
}

type Tester interface {
    DeveloperRole
    GenerateTestCases(code string, testType string) []string
    ExecuteTests(testCases []string) map[string]bool
    GenerateTestReport(results map[string]bool) string
}

type Reviewer interface {
    DeveloperRole
    ReviewCode(code string, standards []string) []string
    CheckSecurity(code string) []string
    EnforceBestPractices(code string) []string
}