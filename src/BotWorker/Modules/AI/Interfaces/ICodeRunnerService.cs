using System.Threading.Tasks;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface ICodeRunnerService
    {
        /// <summary>
        /// 执行系统命令并返回结果
        /// </summary>
        /// <param name="command">命令 (如 dotnet build)</param>
        /// <param name="workingDirectory">工作目录</param>
        /// <returns>执行结果 (包含 ExitCode, Stdout, Stderr)</returns>
        Task<ExecutionResult> ExecuteCommandAsync(string command, string workingDirectory);
    }

    public class ExecutionResult
    {
        public int ExitCode { get; set; }
        public string Stdout { get; set; } = string.Empty;
        public string Stderr { get; set; } = string.Empty;
        public bool Success => ExitCode == 0;
        public string CombinedOutput => $"{Stdout}\n{Stderr}".Trim();
    }
}
