using System;
using System.Diagnostics;
using System.Text;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Interfaces;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.AI.Services
{
    public class CodeRunnerService : ICodeRunnerService
    {
        private readonly ILogger<CodeRunnerService> _logger;

        public CodeRunnerService(ILogger<CodeRunnerService> logger)
        {
            _logger = logger;
        }

        public async Task<ExecutionResult> ExecuteCommandAsync(string command, string workingDirectory)
        {
            _logger.LogInformation("[CodeRunner] Executing command: {Command} in {Dir}", command, workingDirectory);
            
            var result = new ExecutionResult();
            try
            {
                var parts = command.Split(' ', 2);
                var fileName = parts[0];
                var arguments = parts.Length > 1 ? parts[1] : "";

                using var process = new Process();
                process.StartInfo = new ProcessStartInfo
                {
                    FileName = fileName,
                    Arguments = arguments,
                    WorkingDirectory = workingDirectory,
                    RedirectStandardOutput = true,
                    RedirectStandardError = true,
                    UseShellExecute = false,
                    CreateNoWindow = true,
                    StandardOutputEncoding = Encoding.UTF8,
                    StandardErrorEncoding = Encoding.UTF8
                };

                var stdout = new StringBuilder();
                var stderr = new StringBuilder();

                process.OutputDataReceived += (s, e) => { if (e.Data != null) stdout.AppendLine(e.Data); };
                process.ErrorDataReceived += (s, e) => { if (e.Data != null) stderr.AppendLine(e.Data); };

                process.Start();
                process.BeginOutputReadLine();
                process.BeginErrorReadLine();

                await process.WaitForExitAsync();

                result.ExitCode = process.ExitCode;
                result.Stdout = stdout.ToString();
                result.Stderr = stderr.ToString();
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[CodeRunner] Failed to execute command: {Command}", command);
                result.ExitCode = -1;
                result.Stderr = ex.Message;
            }

            return result;
        }
    }
}
