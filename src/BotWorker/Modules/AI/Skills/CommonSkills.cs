using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Interfaces;

namespace BotWorker.Modules.AI.Skills
{
    public class FileSkills : ISkill
    {
        public string Name => "FileTools";
        public string Description => "文件操作工具，支持 LIST, READ, WRITE";
        public string[] SupportedActions => new[] { "LIST", "READ", "WRITE" };

        public async Task<string> ExecuteAsync(string action, string target, string reason, Dictionary<string, string> metadata)
        {
            // 强制要求 ProjectPath，如果没有则使用安全的隔离目录，禁止污染运行目录
            var projectPath = metadata.GetValueOrDefault("ProjectPath");
            if (string.IsNullOrEmpty(projectPath))
            {
                // 如果没有提供 ProjectPath，根据元数据生成一个默认隔离路径
                var tenantId = metadata.GetValueOrDefault("TenantId") ?? "default_tenant";
                var userId = metadata.GetValueOrDefault("UserId") ?? "default_user";
                var taskId = metadata.GetValueOrDefault("TaskId") ?? "unknown_task";
                
                projectPath = Path.Combine(Directory.GetCurrentDirectory(), "BotWorkspaces", tenantId, userId, taskId);
                
                // 可以在这里记录警告，说明技能调用缺少明确的项目路径
            }
            
            if (!Directory.Exists(projectPath))
            {
                Directory.CreateDirectory(projectPath);
            }

            var cmd = action.ToUpper();

            try
            {
                switch (cmd)
                {
                    case "LIST":
                        var files = Directory.GetFiles(projectPath, "*.*", SearchOption.AllDirectories)
                            .Select(f => Path.GetRelativePath(projectPath, f)).ToList();
                        return $"文件列表：\n{string.Join("\n", files)}";

                    case "READ":
                        var readPath = Path.Combine(projectPath, target);
                        if (File.Exists(readPath))
                        {
                            return await File.ReadAllTextAsync(readPath);
                        }
                        return $"错误：文件 {target} 不存在";

                    case "WRITE":
                        var writePath = Path.Combine(projectPath, target);
                        Directory.CreateDirectory(Path.GetDirectoryName(writePath)!);
                        
                        // 优先从元数据中获取 Content，因为 Content 专门用于存储代码或大段文本
                        // 如果没有 Content，则回退使用 reason
                        var content = metadata.GetValueOrDefault("Content") ?? reason;
                        
                        await File.WriteAllTextAsync(writePath, content);
                        return $"已成功写入文件：{target} (长度: {content.Length})";

                    default:
                        return $"FileSkill 不支持行动：{action}";
                }
            }
            catch (Exception ex)
            {
                return $"文件操作失败：{ex.Message}";
            }
        }
    }

    public class ShellSkills : ISkill
    {
        private readonly ICodeRunnerService _codeRunner;

        public ShellSkills(ICodeRunnerService codeRunner)
        {
            _codeRunner = codeRunner;
        }

        public string Name => "ShellTools";
        public string Description => "系统命令工具，支持 BUILD, GIT, COMMAND";
        public string[] SupportedActions => new[] { "BUILD", "GIT", "COMMAND" };

        public async Task<string> ExecuteAsync(string action, string target, string reason, Dictionary<string, string> metadata)
        {
            // 强制要求 ProjectPath，如果没有则使用安全的隔离目录，禁止污染运行目录
            var projectPath = metadata.GetValueOrDefault("ProjectPath");
            if (string.IsNullOrEmpty(projectPath))
            {
                // 如果没有提供 ProjectPath，根据元数据生成一个默认隔离路径
                var tenantId = metadata.GetValueOrDefault("TenantId") ?? "default_tenant";
                var userId = metadata.GetValueOrDefault("UserId") ?? "default_user";
                var taskId = metadata.GetValueOrDefault("TaskId") ?? "unknown_task";
                
                projectPath = Path.Combine(Directory.GetCurrentDirectory(), "BotWorkspaces", tenantId, userId, taskId);
            }

            if (!Directory.Exists(projectPath))
            {
                Directory.CreateDirectory(projectPath);
            }

            var cmd = action.ToUpper();

            try
            {
                switch (cmd)
                {
                    case "BUILD":
                        var buildResult = await _codeRunner.ExecuteCommandAsync("dotnet build", projectPath);
                        return buildResult.Success ? "编译成功" : $"编译失败：\n{buildResult.CombinedOutput}";

                    case "GIT":
                        // target 可能是完整的 git 命令
                        var gitResult = await _codeRunner.ExecuteCommandAsync(target, projectPath);
                        return gitResult.Success ? $"Git 执行成功: {target}" : $"Git 执行失败: {gitResult.CombinedOutput}";

                    case "COMMAND":
                        var cmdResult = await _codeRunner.ExecuteCommandAsync(target, projectPath);
                        return cmdResult.Success ? $"执行成功" : $"执行失败：\n{cmdResult.CombinedOutput}";

                    default:
                        return $"ShellSkill 不支持行动：{action}";
                }
            }
            catch (Exception ex)
            {
                return $"命令执行失败：{ex.Message}";
            }
        }
    }
}
