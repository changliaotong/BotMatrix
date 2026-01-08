using System;
using System.Collections.Generic;
using System.Reflection;
using System.Threading.Tasks;

namespace BotWorker.Domain.Interfaces
{
    // --- 与 Go 项目对齐的进程通信协议模型 ---

    public class Intent
    {
        public string Name { get; set; } = string.Empty;
        public string[] Keywords { get; set; } = Array.Empty<string>();
        public string Regex { get; set; } = string.Empty;
        public int Priority { get; set; }
        public string Action { get; set; } = string.Empty;
        public Dictionary<string, object> Parameters { get; set; } = new();
    }

    public class UIComponent
    {
        public string Type { get; set; } = string.Empty;     // "panel", "button", "tab"
        public string Position { get; set; } = string.Empty; // "sidebar", "dashboard", "chat_action"
        public string Entry { get; set; } = string.Empty;    // URL or HTML file path
        public string Title { get; set; } = string.Empty;
        public string Icon { get; set; } = string.Empty;
    }

    public class EventMessage
    {
        public string ID { get; set; } = Guid.NewGuid().ToString();
        public string Type { get; set; } = "event";
        public string Name { get; set; } = string.Empty;
        public string? CorrelationID { get; set; }
        public Dictionary<string, object> Payload { get; set; } = new();
    }

    public class BotAction
    {
        public string Type { get; set; } = string.Empty;
        public string? Target { get; set; }
        public string? TargetID { get; set; }
        public string? Text { get; set; }
        public string? CorrelationID { get; set; }
        public Dictionary<string, object>? Payload { get; set; }
    }

    public class ResponseMessage
    {
        public string ID { get; set; } = string.Empty;
        public bool OK { get; set; }
        public List<BotAction> Actions { get; set; } = new();
        public string? Error { get; set; }
    }

    /// <summary>
    /// 插件元数据接口
    /// </summary>
    public interface IModuleMetadata
    {
        string Id { get; }
        string Name { get; }
        string Version { get; }
        string Author { get; }
        string Description { get; }
        string Category { get; }
        string[] Permissions { get; }
        string[] Dependencies { get; }
        bool IsEssential { get; }

        // Go 兼容字段
        string[] Events { get; }
    }

    /// <summary>
    /// 插件特性，用于定义插件元数据
    /// </summary>
    [AttributeUsage(AttributeTargets.Class, AllowMultiple = false)]
    public class BotPluginAttribute : Attribute, IModuleMetadata
    {
        public required string Id { get; init; }
        public required string Name { get; init; }
        public string Version { get; init; } = "1.0.0";
        public string Author { get; init; } = "System";
        public string Description { get; init; } = string.Empty;
        public string Category { get; init; } = "General";

        // 扩展字段
        public string[] Permissions { get; init; } = Array.Empty<string>();
        public string[] Dependencies { get; init; } = Array.Empty<string>();
        public bool IsEssential { get; init; } = false; // 是否为核心插件（不可禁用）

        // Go 兼容字段
        public string[] Events { get; init; } = Array.Empty<string>();
    }

    /// <summary>
    /// 机器人插件接口
    /// </summary>
    public interface IPlugin
    {
        /// <summary>
        /// 插件元数据（自动从特性中读取）
        /// </summary>
        IModuleMetadata Metadata => GetType().GetCustomAttribute<BotPluginAttribute>() 
            ?? throw new InvalidOperationException($"Plugin {GetType().Name} is missing BotPluginAttribute");

        /// <summary>
        /// 意图定义 (Go 兼容)
        /// </summary>
        List<Intent> Intents => new();

        /// <summary>
        /// UI 组件定义 (Go 兼容)
        /// </summary>
        List<UIComponent> UI => new();

        /// <summary>
        /// 初始化插件
        /// </summary>
        /// <param name="robot">机器人实例，用于注册技能和事件</param>
        Task InitAsync(IRobot robot);

        /// <summary>
        /// 停止插件
        /// </summary>
        Task StopAsync();
    }
}


