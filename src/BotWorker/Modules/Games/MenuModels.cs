using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Interfaces;

namespace BotWorker.Modules.Games
{
    /// <summary>
    /// 菜单节点类型
    /// </summary>
    public enum MenuNodeType
    {
        Container,  // 容器（包含子菜单）
        Command,    // 指令（执行动作）
        Input,      // 输入（启动多轮对话采集数据）
        Back        // 返回上一级
    }

    /// <summary>
    /// 菜单节点定义
    /// </summary>
    public class MenuNode
    {
        public string Id { get; set; } = string.Empty;
        public string Title { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
        public MenuNodeType Type { get; set; } = MenuNodeType.Container;
        
        // 子菜单列表 (仅 Container 类型有效)
        public List<MenuNode> Children { get; set; } = new();
        
        // 执行动作的技能名 (仅 Command 类型有效)
        public string? ActionSkill { get; set; }
        
        // 多轮对话的问题列表 (仅 Input 类型有效)
        public List<string>? Questions { get; set; }
        
        // 权限要求
        public string[]? Roles { get; set; }
    }

    /// <summary>
    /// 菜单会话状态
    /// </summary>
    public class MenuSession
    {
        public string UserId { get; set; } = string.Empty;
        public List<string> Path { get; set; } = new(); // 菜单路径栈 [root, submenu1, ...]
        public int CurrentQuestionIndex { get; set; } = -1; // -1 表示不在对话中
        public Dictionary<string, string> CollectedData { get; set; } = new();
        public DateTime LastActiveTime { get; set; } = DateTime.Now;

        public string CurrentMenuId => Path.Count > 0 ? Path[Path.Count - 1] : "root";
        
        public bool IsExpired(int timeoutSeconds = 300) 
            => DateTime.Now.Subtract(LastActiveTime).TotalSeconds > timeoutSeconds;
    }
}
