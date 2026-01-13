using System;
using System.Diagnostics;
using System.IO;
using System.Threading.Tasks;
using System.ComponentModel;
using Microsoft.SemanticKernel;
using BotWorker.Modules.AI.Tools;

namespace BotWorker.Modules.AI.Plugins
{
    public class SystemAdminPlugin
    {
        private readonly string _baseDirectory;

        public SystemAdminPlugin()
        {
            _baseDirectory = AppDomain.CurrentDomain.BaseDirectory;
        }

        [KernelFunction(name: "git_status")]
        [Description("获取当前 Git 仓库的状态。")]
        [ToolRisk(ToolRiskLevel.Low, "查看 Git 状态")]
        public async Task<string> GitStatus()
        {
            return await RunGitCommand("status");
        }

        [KernelFunction(name: "git_diff")]
        [Description("查看当前未提交的修改差异。")]
        [ToolRisk(ToolRiskLevel.Low, "查看代码差异")]
        public async Task<string> GitDiff()
        {
            return await RunGitCommand("diff");
        }

        [KernelFunction(name: "git_commit")]
        [Description("提交当前的修改到本地仓库。需要提供提交信息。")]
        [ToolRisk(ToolRiskLevel.High, "提交代码更改")]
        public async Task<string> GitCommit([Description("提交信息")] string message)
        {
            // 确保使用数字员工的身份
            var configResult = await RunGitCommand("config user.name \"Digital Employee\"");
            await RunGitCommand("config user.email \"bot@botmatrix.ai\"");
            
            // 先 add 所有更改
            await RunGitCommand("add .");
            
            return await RunGitCommand($"commit -m \"{message.Replace("\"", "\\\"")}\"");
        }

        [KernelFunction(name: "git_push")]
        [Description("将本地提交推送到远程仓库。属于高风险操作。")]
        [ToolRisk(ToolRiskLevel.High, "推送代码到远程仓库")]
        public async Task<string> GitPush()
        {
            return await RunGitCommand("push");
        }

        [KernelFunction(name: "execute_shell_command")]
        [Description("在系统终端中执行命令（如 dotnet build, npm install 等）。属于极高风险操作。")]
        [ToolRisk(ToolRiskLevel.High, "执行系统终端命令")]
        public async Task<string> ExecuteCommand(
            [Description("要执行的命令")] string command,
            [Description("命令参数")] string arguments = "")
        {
            try
            {
                var startInfo = new ProcessStartInfo
                {
                    FileName = command,
                    Arguments = arguments,
                    RedirectStandardOutput = true,
                    RedirectStandardError = true,
                    UseShellExecute = false,
                    CreateNoWindow = true,
                    WorkingDirectory = _baseDirectory
                };

                using var process = Process.Start(startInfo);
                if (process == null) return "Error: 无法启动进程。";

                var output = await process.StandardOutput.ReadToEndAsync();
                var error = await process.StandardError.ReadToEndAsync();
                await process.WaitForExitAsync();

                if (process.ExitCode != 0)
                {
                    return $"Error (ExitCode {process.ExitCode}):\n{error}\nOutput:\n{output}";
                }

                return output;
            }
            catch (Exception ex)
            {
                return $"Error: 执行命令失败 - {ex.Message}";
            }
        }

        private async Task<string> RunGitCommand(string args)
        {
            try
            {
                var startInfo = new ProcessStartInfo
                {
                    FileName = "git",
                    Arguments = args,
                    RedirectStandardOutput = true,
                    RedirectStandardError = true,
                    UseShellExecute = false,
                    CreateNoWindow = true,
                    WorkingDirectory = _baseDirectory
                };

                using var process = Process.Start(startInfo);
                if (process == null) return "Error: 无法启动 Git 进程。";

                var output = await process.StandardOutput.ReadToEndAsync();
                var error = await process.StandardError.ReadToEndAsync();
                await process.WaitForExitAsync();

                if (process.ExitCode != 0)
                {
                    return $"Git Error (ExitCode {process.ExitCode}): {error}";
                }

                return output;
            }
            catch (Exception ex)
            {
                return $"Error: 执行 Git 命令失败 - {ex.Message}";
            }
        }
    }
}
