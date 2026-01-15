using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Interfaces;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Modules.AI.Tools
{
    /// <summary>
    /// 工具调用审计日志
    /// </summary>
    public class ToolAuditLog
    {
        public long Id { get; set; }
        public string Guid { get; set; } = string.Empty;
        public string TaskId { get; set; } = string.Empty;
        public string StaffId { get; set; } = string.Empty;
        public string ToolName { get; set; } = string.Empty;
        public string InputArgs { get; set; } = string.Empty; // JSON
        public string OutputResult { get; set; } = string.Empty; // JSON
        public ToolRiskLevel RiskLevel { get; set; }
        public string Status { get; set; } = "Success"; // Success, Failed, PendingApproval, Approved, Rejected
        public string ApprovedBy { get; set; } = string.Empty; // 人工审批者
        public string RejectionReason { get; set; } = string.Empty; // 拒绝原因
        public DateTime? ApprovedAt { get; set; }
        public DateTime CreateTime { get; set; } = DateTime.Now;
    }

    public interface IToolAuditService
    {
        Task<string> LogCallAsync(string taskId, string staffId, string toolName, string input, ToolRiskLevel risk);
        Task UpdateResultAsync(string logId, string output, string status = "Success");
        Task MarkAsPendingApprovalAsync(string logId);
        Task ApproveAsync(string logId, string approver);
        Task RejectAsync(string logId, string approver, string reason);
        Task<IEnumerable<ToolAuditLog>> GetPendingApprovalsAsync();
    }

    public class ToolAuditService : IToolAuditService
    {
        private readonly IServiceProvider _serviceProvider;

        public ToolAuditService(IServiceProvider serviceProvider)
        {
            _serviceProvider = serviceProvider;
        }

        private IToolAuditRepository GetRepository()
        {
            return _serviceProvider.GetRequiredService<IToolAuditRepository>();
        }

        public async Task<string> LogCallAsync(string taskId, string staffId, string toolName, string input, ToolRiskLevel risk)
        {
            try
            {
                var log = new ToolAuditLog
                {
                    Guid = Guid.NewGuid().ToString(),
                    TaskId = taskId,
                    StaffId = staffId,
                    ToolName = toolName,
                    InputArgs = input,
                    RiskLevel = risk,
                    Status = "InProgress",
                    CreateTime = DateTime.Now
                };
                await GetRepository().AddAsync(log);
                return log.Guid;
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[ToolAuditService] Error logging call: {ex.Message}");
                return Guid.NewGuid().ToString(); // 返回一个新的 GUID，即使保存失败也允许流程继续
            }
        }

        public async Task UpdateResultAsync(string logId, string output, string status = "Success")
        {
            try
            {
                var repo = GetRepository();
                var log = await repo.GetByGuidAsync(logId);
                if (log != null)
                {
                    log.OutputResult = output;
                    log.Status = status;
                    await repo.UpdateAsync(log);
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[ToolAuditService] Error updating result: {ex.Message}");
            }
        }

        public async Task MarkAsPendingApprovalAsync(string logId)
        {
            var repo = GetRepository();
            var log = await repo.GetByGuidAsync(logId);
            if (log != null)
            {
                log.Status = "PendingApproval";
                await repo.UpdateAsync(log);
            }
        }

        public async Task ApproveAsync(string logId, string approver)
        {
            var repo = GetRepository();
            var log = await repo.GetByGuidAsync(logId);
            if (log != null)
            {
                log.Status = "Approved";
                log.ApprovedBy = approver;
                log.ApprovedAt = DateTime.Now;
                await repo.UpdateAsync(log);
            }
        }

        public async Task RejectAsync(string logId, string approver, string reason)
        {
            var repo = GetRepository();
            var log = await repo.GetByGuidAsync(logId);
            if (log != null)
            {
                log.Status = "Rejected";
                log.ApprovedBy = approver;
                log.RejectionReason = reason;
                log.ApprovedAt = DateTime.Now;
                await repo.UpdateAsync(log);
            }
        }

        public async Task<IEnumerable<ToolAuditLog>> GetPendingApprovalsAsync()
        {
            return await GetRepository().GetPendingApprovalsAsync();
        }
    }
}
