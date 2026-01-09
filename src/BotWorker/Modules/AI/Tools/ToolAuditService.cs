using System;
using System.Threading.Tasks;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Modules.AI.Tools
{
    /// <summary>
    /// 工具调用审计日志
    /// </summary>
    public class ToolAuditLog : MetaDataGuid<ToolAuditLog>
    {
        public override string TableName => "ToolAuditLogs";
        public override string KeyField => "Id";

        public string TaskId { get; set; } = string.Empty;
        public string StaffId { get; set; } = string.Empty;
        public string ToolName { get; set; } = string.Empty;
        public string InputArgs { get; set; } = string.Empty; // JSON
        public string OutputResult { get; set; } = string.Empty; // JSON
        public ToolRiskLevel RiskLevel { get; set; }
        public string Status { get; set; } = "Success"; // Success, Failed, PendingApproval
        public string ApprovedBy { get; set; } = string.Empty; // 人工审批者
        public DateTime CreateTime { get; set; } = DateTime.Now;
    }

    public interface IToolAuditService
    {
        Task<string> LogCallAsync(string taskId, string staffId, string toolName, string input, ToolRiskLevel risk);
        Task UpdateResultAsync(string logId, string output, string status = "Success");
        Task MarkAsPendingApprovalAsync(string logId);
        Task ApproveAsync(string logId, string approver);
    }

    public class ToolAuditService : IToolAuditService
    {
        public async Task<string> LogCallAsync(string taskId, string staffId, string toolName, string input, ToolRiskLevel risk)
        {
            var log = new ToolAuditLog
            {
                TaskId = taskId,
                StaffId = staffId,
                ToolName = toolName,
                InputArgs = input,
                RiskLevel = risk,
                Status = "InProgress"
            };
            await log.InsertAsync();
            return log.Guid.ToString();
        }

        public async Task UpdateResultAsync(string logId, string output, string status = "Success")
        {
            var log = await ToolAuditLog.GetByGuidAsync(logId);
            if (log != null)
            {
                log.OutputResult = output;
                log.Status = status;
                await log.UpdateAsync();
            }
        }

        public async Task MarkAsPendingApprovalAsync(string logId)
        {
            var log = await ToolAuditLog.GetByGuidAsync(logId);
            if (log != null)
            {
                log.Status = "PendingApproval";
                await log.UpdateAsync();
            }
        }

        public async Task ApproveAsync(string logId, string approver)
        {
            var log = await ToolAuditLog.GetByGuidAsync(logId);
            if (log != null)
            {
                log.Status = "Approved";
                log.ApprovedBy = approver;
                await log.UpdateAsync();
            }
        }
    }
}
