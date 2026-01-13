using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Modules.AI.Models.Evolution
{
    public class EmployeeInstance : MetaDataGuid<EmployeeInstance>
    {
        public override string TableName => "EmployeeInstance";
        public override string KeyField => "Id";

        public string EmployeeId { get; set; } = string.Empty; // 员工唯一标识
        public string JobId { get; set; } = string.Empty; // 关联的岗位 JobId
        
        public string SkillSet { get; set; } = "[]"; // JSON Array, 绑定的技能列表
        public string PermissionSet { get; set; } = "[]"; // JSON Array, 权限列表
        
        public string State { get; set; } = "Idle"; // Idle, Working, Paused
        public int Version { get; set; } = 1;
        public DateTime CreatedAt { get; set; } = DateTime.Now;
    }
}
