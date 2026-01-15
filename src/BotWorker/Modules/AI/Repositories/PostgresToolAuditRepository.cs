using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Infrastructure.Persistence.Repositories;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Tools;
using Dapper;

namespace BotWorker.Modules.AI.Repositories
{
    public class PostgresToolAuditRepository : BasePostgresRepository<ToolAuditLog>, IToolAuditRepository
    {
        public PostgresToolAuditRepository(string? connectionString = null) 
            : base("ai_tool_audit_logs", connectionString)
        {
        }

        public async Task<long> AddAsync(ToolAuditLog log)
        {
            const string sql = @"
                INSERT INTO ai_tool_audit_logs (
                    guid, task_id, staff_id, tool_name, input_args, output_result, 
                    risk_level, status, approved_by, rejection_reason, approved_at, create_time
                ) VALUES (
                    @Guid, @TaskId, @StaffId, @ToolName, @InputArgs, @OutputResult, 
                    @RiskLevel, @Status, @ApprovedBy, @RejectionReason, @ApprovedAt, @CreateTime
                ) RETURNING id";
            
            using var conn = CreateConnection();
            if (string.IsNullOrEmpty(log.Guid)) log.Guid = Guid.NewGuid().ToString();
            return await conn.ExecuteScalarAsync<long>(sql, log);
        }

        public async Task<bool> UpdateAsync(ToolAuditLog log)
        {
            const string sql = @"
                UPDATE ai_tool_audit_logs SET 
                    output_result = @OutputResult, 
                    status = @Status, 
                    approved_by = @ApprovedBy, 
                    rejection_reason = @RejectionReason, 
                    approved_at = @ApprovedAt,
                    updated_at = CURRENT_TIMESTAMP
                WHERE guid = @Guid";
            
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, log) > 0;
        }

        public async Task<ToolAuditLog?> GetByGuidAsync(string guid)
        {
            const string sql = "SELECT * FROM ai_tool_audit_logs WHERE guid = @guid";
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<ToolAuditLog>(sql, new { guid });
        }

        public async Task<IEnumerable<ToolAuditLog>> GetPendingApprovalsAsync()
        {
            const string sql = "SELECT * FROM ai_tool_audit_logs WHERE status = 'PendingApproval' ORDER BY create_time DESC";
            using var conn = CreateConnection();
            return await conn.QueryAsync<ToolAuditLog>(sql);
        }
    }
}
