using System;
using System.Threading.Tasks;
using BotWorker.Bots.BotMessages;

namespace BotWorker.Core.Commands
{
    /// <summary>
    /// 指令执行结果
    /// </summary>
    public class CommandResult
    {
        public bool Success { get; set; }
        public string Message { get; set; }
        public bool Intercept { get; set; } = true; // 是否拦截后续处理

        public static CommandResult Intercepted(string message = "") => new() { Success = true, Message = message, Intercept = true };
        public static CommandResult Continue() => new() { Success = true, Intercept = false };
    }

    /// <summary>
    /// 机器人指令接�?    /// </summary>
    public interface IBotCommand
    {
        /// <summary>
        /// 指令匹配规则（正则表达式或关键词�?        /// </summary>
        string MatchPattern { get; }

        /// <summary>
        /// 执行指令
        /// </summary>
        Task<CommandResult> ExecuteAsync(BotMessage botMsg);
    }
}


