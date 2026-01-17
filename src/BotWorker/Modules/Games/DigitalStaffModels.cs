using System;
using Dapper.Contrib.Extensions;

namespace BotWorker.Modules.Games
{
    public enum StaffRole
    {
        ProductManager, // 需求分析与规划
        Developer,      // 自动编程与系统升级
        Tester,         // 自动化测试与质量控制
        CustomerService,// 自动答疑与用户引导
        Sales,          // 自动营销与流量变现
        AfterSales      // 异常监测与系统维护
    }

    [Table("digital_staff")]
    public class DigitalStaff
    {
        [Key]
        public long Id { get; set; }
        public Guid Guid { get; set; } = Guid.NewGuid();
        public string OwnerUserId { get; set; } = string.Empty;
        public string StaffName { get; set; } = string.Empty;
        public StaffRole Role { get; set; }
        public int Level { get; set; } = 1;
        public long TotalProfitGenerated { get; set; } = 0;
        public long SalaryToken { get; set; } = 0;
        public long SalaryLimit { get; set; } = 1000000;
        public double KpiScore { get; set; } = 100.0;
        public string SystemPrompt { get; set; } = string.Empty;
        public DateTime HireDate { get; set; } = DateTime.Now;
        public string CurrentStatus { get; set; } = "Idle";
        public string AssignedTaskId { get; set; } = string.Empty;
    }

    [Table("CognitiveMemories")]
    public class CognitiveMemory
    {
        [Key]
        public long Id { get; set; }
        public Guid Guid { get; set; } = Guid.NewGuid();
        public string StaffId { get; set; } = string.Empty;
        public string UserId { get; set; } = string.Empty;
        public string Category { get; set; } = "General";
        public string Content { get; set; } = string.Empty;
        public int Importance { get; set; } = 3;
        public string Embedding { get; set; } = string.Empty;
        public DateTime LastSeen { get; set; } = DateTime.Now;
        public DateTime CreateTime { get; set; } = DateTime.Now;
    }

    [Table("staff_kpis")]
    public class StaffKpi
    {
        [Key]
        public long Id { get; set; }
        public Guid Guid { get; set; } = Guid.NewGuid();
        public string StaffId { get; set; } = string.Empty;
        public string MetricName { get; set; } = string.Empty;
        public double Score { get; set; } = 0;
        public string Detail { get; set; } = string.Empty;
        public DateTime CreateTime { get; set; } = DateTime.Now;
    }

    [Table("staff_tasks")]
    public class StaffTask
    {
        [Key]
        public long Id { get; set; }
        public Guid Guid { get; set; } = Guid.NewGuid();
        public string Title { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
        public string TaskType { get; set; } = string.Empty;
        public string Status { get; set; } = "Pending";
        public string CreatorUserId { get; set; } = string.Empty;
        public string ExecutorStaffId { get; set; } = string.Empty;
        public string Result { get; set; } = string.Empty;
        public DateTime CreateTime { get; set; } = DateTime.Now;
        public DateTime? CompleteTime { get; set; }
    }
}
