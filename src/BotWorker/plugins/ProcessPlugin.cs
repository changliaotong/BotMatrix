using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Text.Json;
using System.Threading.Tasks;
using BotWorker.Domain.Interfaces;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.Plugins
{
    /// <summary>
    /// 进程插件适配器：与 Go 项目插件系统完全一致，通过 Stdin/Stdout 通信
    /// </summary>
    public class ProcessPlugin : IPlugin
    {
        private readonly string _executablePath;
        private readonly string _workingDirectory;
        private readonly IModuleMetadata _metadata;
        private Process? _process;
        private StreamWriter? _stdin;
        private readonly Dictionary<string, TaskCompletionSource<ResponseMessage>> _pendingRequests = new();
        private readonly ILogger _logger;
        private IRobot? _robot;

        public IModuleMetadata Metadata => _metadata;

        public ProcessPlugin(IModuleMetadata metadata, string executablePath, ILogger logger, string? workingDirectory = null)
        {
            _metadata = metadata;
            _executablePath = executablePath;
            _logger = logger;
            _workingDirectory = workingDirectory ?? Path.GetDirectoryName(executablePath) ?? "";
        }

        public async Task InitAsync(IRobot robot)
        {
            _robot = robot;
            StartProcess();

            // 1. 根据插件配置的 Intents 注册技能
            if (_metadata is PluginConfig config && config.Intents != null)
            { 
                foreach (var intent in config.Intents)
                {
                    await robot.RegisterSkillAsync(new SkillCapability
                    {
                        Name = intent.Name,
                        Commands = intent.Keywords,
                        Usage = intent.Regex // 暂时将正则存放在 Usage 字段
                    }, async (ctx, args) =>
                    {
                        // 转发指令事件
                        var response = await SendRequestAsync("on_command", new Dictionary<string, object>
                        {
                            { "command", intent.Name },
                            { "args", args },
                            { "text", ctx.RawMessage },
                            { "from", ctx.UserId },
                            { "group_id", ctx.GroupId ?? "" }
                        });
                        
                        // 提取响应中的第一个 text action 作为返回
                        return response.Actions?.FirstOrDefault(a => a.Type == "reply" || a.Type == "send_message")?.Text ?? "";
                    });
                }
            }

            // 2. 监听并转发通用消息事件
            await robot.RegisterEventAsync("message", async (ctx) =>
            {
                await SendEventAsync("on_message", new Dictionary<string, object>
                {
                    { "text", ctx.RawMessage },
                    { "from", ctx.UserId },
                    { "group_id", ctx.GroupId ?? "" },
                    { "platform", ctx.Platform },
                    { "bot_id", ctx.BotId }
                });
            });
        }

        private async Task<ResponseMessage> SendRequestAsync(string eventName, Dictionary<string, object> payload)
        {
            var requestId = Guid.NewGuid().ToString();
            var tcs = new TaskCompletionSource<ResponseMessage>();
            _pendingRequests[requestId] = tcs;

            await SendEventAsync(eventName, payload, requestId);

            // 设置超时，防止插件挂死
            var completedTask = await Task.WhenAny(tcs.Task, Task.Delay(5000));
            if (completedTask == tcs.Task)
            {
                return await tcs.Task;
            }
            
            _pendingRequests.Remove(requestId);
            return new ResponseMessage { OK = false, Error = "Request timeout" };
        }

        private void StartProcess()
        {
            var startInfo = new ProcessStartInfo
            {
                FileName = _executablePath,
                WorkingDirectory = _workingDirectory,
                UseShellExecute = false,
                RedirectStandardInput = true,
                RedirectStandardOutput = true,
                RedirectStandardError = true,
                CreateNoWindow = true
            };

            _process = new Process { StartInfo = startInfo };
            _process.OutputDataReceived += (s, e) => HandleStdout(e.Data);
            _process.ErrorDataReceived += (s, e) => {
                if (!string.IsNullOrEmpty(e.Data)) Console.WriteLine($"[Plugin:{_metadata.Name}] ERR: {e.Data}");
            };

            _process.Start();
            _stdin = _process.StandardInput;
            _process.BeginOutputReadLine();
            _process.BeginErrorReadLine();
        }

        private async void HandleStdout(string? data)
        {
            if (string.IsNullOrEmpty(data)) return;

            try
            {
                var response = JsonSerializer.Deserialize<ResponseMessage>(data, new JsonSerializerOptions { PropertyNameCaseInsensitive = true });
                if (response == null) return;

                // 1. 处理异步 Actions
                if (response.Actions != null)
                {
                    foreach (var action in response.Actions)
                    {
                        _ = ExecuteActionAsync(action); // 异步执行，不阻塞 stdout 读取
                    }
                }

                // 2. 如果是针对某个请求的响应（通过 ID 匹配）
                if (!string.IsNullOrEmpty(response.ID) && _pendingRequests.TryGetValue(response.ID, out var tcs))
                {
                    tcs.SetResult(response);
                    _pendingRequests.Remove(response.ID);
                }
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "[ProcessPlugin:{Name}] Error parsing stdout: {Data}", _metadata.Name, data);
            }
        }

        private async Task ExecuteActionAsync(BotAction action)
        {
            if (_robot == null) return;

            try
            {
                switch (action.Type)
                {
                    case "send_message":
                    case "reply":
                    case "send_text":
                        var text = action.Text ?? action.Payload?["text"]?.ToString();
                        if (!string.IsNullOrEmpty(text))
                        {
                            // 优先使用 TargetID，如果没有则根据上下文（通常在 Payload 中）
                            var targetId = action.TargetID ?? action.Payload?["user_id"]?.ToString() ?? "";
                            var groupId = action.Target == "group" ? action.TargetID : (action.Payload?.ContainsKey("group_id") == true ? action.Payload["group_id"]?.ToString() : null);
                            
                            await _robot.SendMessageAsync("onebot", "", groupId, targetId, text);
                        }
                        break;
                    
                    case "call_skill":
                        if (action.Payload != null && action.Payload.TryGetValue("skill", out var skillObj))
                        {
                            var skillName = skillObj.ToString() ?? "";
                            var skillArgs = action.Payload.ContainsKey("args") ? 
                                JsonSerializer.Deserialize<string[]>(action.Payload["args"].ToString() ?? "[]") : 
                                Array.Empty<string>();
                            
                            // 构造上下文（这里可能需要更完善的上下文恢复逻辑）
                            var ev = new BotWorker.Infrastructure.Communication.OneBot.OneBotEvent { 
                                UserIdLong = action.Payload.ContainsKey("from") ? long.Parse(action.Payload["from"].ToString() ?? "0") : 0,
                                GroupIdLong = action.Payload.ContainsKey("group_id") ? long.Parse(action.Payload["group_id"].ToString() ?? "0") : 0
                            };
                            var ctx = new PluginContext(ev, "onebot", "", null!, null!, _logger!);

                            var result = await _robot.CallSkillAsync(skillName, ctx, skillArgs!);
                            
                            // 如果有 correlation_id，则将结果发回插件
                            if (!string.IsNullOrEmpty(action.CorrelationID))
                            {
                                await SendEventAsync("skill_result", new Dictionary<string, object>
                                {
                                    { "result", result },
                                    { "skill", skillName }
                                }, action.CorrelationID);
                            }
                        }
                        break;
                }
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "[ProcessPlugin:{Name}] Error executing action {Type}", _metadata.Name, action.Type);
            }
        }

        public async Task SendEventAsync(string eventName, Dictionary<string, object> payload, string? correlationId = null)
        {
            if (_stdin == null) return;

            var msg = new EventMessage
            {
                Name = eventName,
                Payload = payload,
                CorrelationID = correlationId
            };

            var json = JsonSerializer.Serialize(msg);
            await _stdin.WriteLineAsync(json);
            await _stdin.FlushAsync();
        }

        public async Task StopAsync()
        {
            try
            {
                if (_process != null && !_process.HasExited)
                {
                    _process.Kill();
                    await _process.WaitForExitAsync();
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[ProcessPlugin] Error stopping process: {ex.Message}");
            }
        }
    }
}