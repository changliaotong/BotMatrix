using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models.Evolution;
using Dapper;
using Npgsql;

using BotWorker.Infrastructure.Persistence.Repositories;

namespace BotWorker.Modules.AI.Repositories
{
    public class PostgresJobDefinitionRepository : BasePostgresRepository<JobDefinition>, IJobDefinitionRepository
    {
        public PostgresJobDefinitionRepository(string? connectionString = null) 
            : base("ai_job_definitions", connectionString)
        {
        }

        public async Task<JobDefinition?> GetByKeyAsync(string jobKey)
        {
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<JobDefinition>(
                $"SELECT * FROM {_tableName} WHERE job_key = @jobKey", new { jobKey });
        }

        public async Task<IEnumerable<JobDefinition>> GetActiveJobsAsync()
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<JobDefinition>($"SELECT * FROM {_tableName} WHERE is_active = true ORDER BY id DESC");
        }

        public async Task<long> AddAsync(JobDefinition entity)
        {
            const string sql = @"
                INSERT INTO ai_job_definitions (
                    job_key, name, purpose, inputs_schema, outputs_schema, constraints, 
                    system_prompt, tool_schema, workflow, model_selection_strategy, version, is_active
                ) VALUES (
                    @JobKey, @Name, @Purpose, @InputsSchema::jsonb, @OutputsSchema::jsonb, @Constraints::jsonb, 
                    @SystemPrompt, @ToolSchema::jsonb, @Workflow::jsonb, @ModelSelectionStrategy, @Version, @IsActive
                ) RETURNING id";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, entity);
        }

        public async Task<bool> UpdateAsync(JobDefinition entity)
        {
            const string sql = @"
                UPDATE ai_job_definitions SET 
                    name = @Name, purpose = @Purpose, inputs_schema = @InputsSchema::jsonb, 
                    outputs_schema = @OutputsSchema::jsonb, constraints = @Constraints::jsonb, 
                    system_prompt = @SystemPrompt, tool_schema = @ToolSchema::jsonb,
                    workflow = @Workflow::jsonb, model_selection_strategy = @ModelSelectionStrategy, 
                    version = @Version, is_active = @IsActive
                WHERE id = @Id";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, entity) > 0;
        }
    }

    public class PostgresSkillDefinitionRepository : BasePostgresRepository<SkillDefinition>, ISkillDefinitionRepository
    {
        public PostgresSkillDefinitionRepository(string? connectionString = null)
            : base("ai_skill_definitions", connectionString)
        {
        }

        public async Task<SkillDefinition?> GetByKeyAsync(string skillKey)
        {
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<SkillDefinition>(
                $"SELECT * FROM {_tableName} WHERE skill_key = @skillKey", new { skillKey });
        }

        public async Task<IEnumerable<SkillDefinition>> GetByKeysAsync(IEnumerable<string> skillKeys)
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<SkillDefinition>(
                $"SELECT * FROM {_tableName} WHERE skill_key = ANY(@skillKeys)", new { skillKeys = skillKeys.ToArray() });
        }

        public async Task<IEnumerable<SkillDefinition>> GetByActionAsync(string actionName)
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<SkillDefinition>(
                $"SELECT * FROM {_tableName} WHERE action_name = @actionName", new { actionName });
        }

        public async Task<long> AddAsync(SkillDefinition entity)
        {
            const string sql = @"
                INSERT INTO ai_skill_definitions (
                    skill_key, name, description, action_name, parameter_schema, is_builtin, script_content
                ) VALUES (
                    @SkillKey, @Name, @Description, @ActionName, @ParameterSchema::jsonb, @IsBuiltin, @ScriptContent
                ) RETURNING id";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, entity);
        }

        public async Task<bool> UpdateAsync(SkillDefinition entity)
        {
            const string sql = @"
                UPDATE ai_skill_definitions SET 
                    name = @Name, description = @Description, action_name = @ActionName, 
                    parameter_schema = @ParameterSchema::jsonb, is_builtin = @IsBuiltin, 
                    script_content = @ScriptContent
                WHERE id = @Id";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, entity) > 0;
        }
    }

    public class PostgresEmployeeInstanceRepository : BasePostgresRepository<EmployeeInstance>, IEmployeeInstanceRepository
    {
        public PostgresEmployeeInstanceRepository(string? connectionString = null)
            : base("ai_employee_instances", connectionString)
        {
        }

        public async Task<EmployeeInstance?> GetByEmployeeIdAsync(string employeeId)
        {
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<EmployeeInstance>(
                $"SELECT * FROM {_tableName} WHERE employee_id = @employeeId", new { employeeId });
        }

        public async Task<IEnumerable<EmployeeInstance>> GetByBotIdAsync(string botId)
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<EmployeeInstance>($"SELECT * FROM {_tableName} WHERE bot_id = @botId", new { botId });
        }

        public async Task<IEnumerable<EmployeeInstance>> GetByJobIdAsync(long jobId)
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<EmployeeInstance>($"SELECT * FROM {_tableName} WHERE job_id = @jobId", new { jobId });
        }

        public async Task<long> AddAsync(EmployeeInstance entity)
        {
            const string sql = @"
                INSERT INTO ai_employee_instances (
                    employee_id, bot_id, agent_id, job_id, name, title, department, online_status, state, 
                    salary_token_used, salary_token_limit, kpi_score, experience_data
                ) VALUES (
                    @EmployeeId, @BotId, @AgentId, @JobId, @Name, @Title, @Department, @OnlineStatus, @State, 
                    @SalaryTokenUsed, @SalaryTokenLimit, @KpiScore, @ExperienceData::jsonb
                ) RETURNING id";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, entity);
        }

        public async Task<bool> UpdateAsync(EmployeeInstance entity)
        {
            const string sql = @"
                UPDATE ai_employee_instances SET 
                    agent_id = @AgentId, job_id = @JobId, name = @Name, title = @Title, 
                    department = @Department, online_status = @OnlineStatus, state = @State, 
                    salary_token_used = @SalaryTokenUsed, salary_token_limit = @SalaryTokenLimit, 
                    kpi_score = @KpiScore, experience_data = @ExperienceData::jsonb
                WHERE id = @Id";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, entity) > 0;
        }

        public async Task<bool> UpdateStatusAsync(long id, string onlineStatus, string state)
        {
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(
                $"UPDATE {_tableName} SET online_status = @onlineStatus, state = @state WHERE id = @id", 
                new { id, onlineStatus, state }) > 0;
        }
    }

    public class PostgresTaskRecordRepository : BasePostgresRepository<TaskRecord>, ITaskRecordRepository
    {
        public PostgresTaskRecordRepository(string? connectionString = null)
            : base("ai_task_records", connectionString)
        {
        }

        public async Task<TaskRecord?> GetByExecutionIdAsync(Guid executionId)
        {
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<TaskRecord>(
                $"SELECT * FROM {_tableName} WHERE execution_id = @executionId", new { executionId });
        }

        public async Task<IEnumerable<TaskRecord>> GetByAssigneeIdAsync(long assigneeId)
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<TaskRecord>($"SELECT * FROM {_tableName} WHERE assignee_id = @assigneeId ORDER BY id DESC", new { assigneeId });
        }

        public async Task<IEnumerable<TaskRecord>> GetRecentAsync(int limit)
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<TaskRecord>($"SELECT * FROM {_tableName} ORDER BY id DESC LIMIT @limit", new { limit });
        }

        public async Task<IEnumerable<TaskRecord>> GetByParentIdAsync(long parentId)
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<TaskRecord>($"SELECT * FROM {_tableName} WHERE parent_task_id = @parentId ORDER BY id ASC", new { parentId });
        }

        public async Task<long> AddAsync(TaskRecord entity)
        {
            const string sql = @"
                INSERT INTO ai_task_records (
                    execution_id, title, description, initiator_id, assignee_id, status, progress, 
                    plan_data, result_data, parent_task_id, started_at, finished_at
                ) VALUES (
                    @ExecutionId, @Title, @Description, @InitiatorId, @AssigneeId, @Status, @Progress, 
                    @PlanData::jsonb, @ResultData::jsonb, @ParentTaskId, @StartedAt, @FinishedAt
                ) RETURNING id";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, entity);
        }

        public async Task<bool> UpdateAsync(TaskRecord entity)
        {
            const string sql = @"
                UPDATE ai_task_records SET 
                    status = @Status, progress = @Progress, plan_data = @PlanData::jsonb, 
                    result_data = @ResultData::jsonb, started_at = @StartedAt, finished_at = @FinishedAt
                WHERE id = @Id";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, entity) > 0;
        }

        public async Task<bool> UpdateStatusAsync(long id, string status, int progress)
        {
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(
                $"UPDATE {_tableName} SET status = @status, progress = @progress WHERE id = @id", 
                new { id, status, progress }) > 0;
        }
    }

    public class PostgresTaskStepRepository : BasePostgresRepository<TaskStep>, ITaskStepRepository
    {
        public PostgresTaskStepRepository(string? connectionString = null)
            : base("ai_task_steps", connectionString)
        {
        }

        public async Task<IEnumerable<TaskStep>> GetByTaskIdAsync(long taskId)
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<TaskStep>($"SELECT * FROM {_tableName} WHERE task_id = @taskId ORDER BY step_index ASC", new { taskId });
        }

        public async Task<long> AddAsync(TaskStep entity)
        {
            const string sql = @"
                INSERT INTO ai_task_steps (
                    task_id, step_index, name, input_data, output_data, status, duration_ms, error_message
                ) VALUES (
                    @TaskId, @StepIndex, @Name, @InputData::jsonb, @OutputData::jsonb, @Status, @DurationMs, @ErrorMessage
                ) RETURNING id";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, entity);
        }

        public async Task<bool> UpdateAsync(TaskStep entity)
        {
            const string sql = @"
                UPDATE ai_task_steps SET 
                    status = @Status, output_data = @OutputData::jsonb, 
                    duration_ms = @DurationMs, error_message = @ErrorMessage
                WHERE id = @Id";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, entity) > 0;
        }
    }
}
