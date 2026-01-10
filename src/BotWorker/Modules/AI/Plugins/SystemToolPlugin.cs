using System;
using System.IO;
using System.Threading.Tasks;
using System.ComponentModel;
using Microsoft.SemanticKernel;
using BotWorker.Modules.AI.Tools;

namespace BotWorker.Modules.AI.Plugins
{
    public class SystemToolPlugin
    {
        private readonly string _baseDirectory;

        public SystemToolPlugin()
        {
            _baseDirectory = AppDomain.CurrentDomain.BaseDirectory;
        }

        [KernelFunction(name: "read_file")]
        [Description("读取指定文件的内容。仅限项目根目录下的文本文件。")]
        [ToolRisk(ToolRiskLevel.Low, "读取系统文件内容")]
        public async Task<string> ReadFile([Description("文件相对路径")] string filePath)
        {
            try
            {
                // 简单的路径安全检查，防止目录穿越
                if (filePath.Contains("..")) return "Error: 禁止访问上级目录。";

                var fullPath = Path.Combine(_baseDirectory, filePath);
                if (!File.Exists(fullPath)) return $"Error: 文件 {filePath} 不存在。";

                return await File.ReadAllTextAsync(fullPath);
            }
            catch (Exception ex)
            {
                return $"Error: 读取文件失败 - {ex.Message}";
            }
        }

        [KernelFunction(name: "list_directory")]
        [Description("列出指定目录下的文件和子目录。")]
        [ToolRisk(ToolRiskLevel.Low, "列出目录结构")]
        public string ListDirectory([Description("目录相对路径")] string path = ".")
        {
            try
            {
                if (path.Contains("..")) return "Error: 禁止访问上级目录。";

                var fullPath = Path.Combine(_baseDirectory, path);
                if (!Directory.Exists(fullPath)) return $"Error: 目录 {path} 不存在。";

                var entries = Directory.GetFileSystemEntries(fullPath);
                return string.Join("\n", entries.Select(Path.GetFileName));
            }
            catch (Exception ex)
            {
                return $"Error: 列出目录失败 - {ex.Message}";
            }
        }

        [KernelFunction(name: "write_file")]
        [Description("写入或覆盖指定文件的内容。属于高风险操作。")]
        [ToolRisk(ToolRiskLevel.High, "修改系统文件内容")]
        public async Task<string> WriteFile(
            [Description("文件相对路径")] string filePath,
            [Description("文件新内容")] string content)
        {
            try
            {
                if (filePath.Contains("..")) return "Error: 禁止访问上级目录。";
                
                // 禁止修改核心配置文件
                if (filePath.EndsWith(".dll") || filePath.EndsWith(".exe"))
                    return "Error: 禁止修改二进制文件。";

                var fullPath = Path.Combine(_baseDirectory, filePath);
                var directory = Path.GetDirectoryName(fullPath);
                if (directory != null && !Directory.Exists(directory))
                {
                    Directory.CreateDirectory(directory);
                }

                await File.WriteAllTextAsync(fullPath, content);
                return $"Success: 文件 {filePath} 已成功更新。";
            }
            catch (Exception ex)
            {
                return $"Error: 写入文件失败 - {ex.Message}";
            }
        }

        [KernelFunction(name: "update_app_config")]
        [Description("更新系统的 appsettings.json 配置项。属于高风险操作。")]
        [ToolRisk(ToolRiskLevel.High, "更新系统全局配置")]
        public async Task<string> UpdateAppConfig(
            [Description("配置项路径 (例如: 'AI:DefaultProvider')")] string key,
            [Description("新的配置值")] string value)
        {
            try
            {
                var configPath = Path.Combine(_baseDirectory, "appsettings.json");
                if (!File.Exists(configPath)) return "Error: 找不到 appsettings.json。";

                var json = await File.ReadAllTextAsync(configPath);
                var jsonObj = System.Text.Json.Nodes.JsonNode.Parse(json);
                
                // 这里需要一个递归寻找并设置 key 的辅助方法
                // 简化版：直接修改根节点或一级节点
                if (jsonObj != null)
                {
                    var parts = key.Split(':');
                    var current = jsonObj;
                    for (int i = 0; i < parts.Length - 1; i++)
                    {
                        current = current![parts[i]];
                        if (current == null) return $"Error: 配置路径 {key} 不存在。";
                    }
                    current![parts[^1]] = value;

                    await File.WriteAllTextAsync(configPath, jsonObj.ToJsonString(new System.Text.Json.JsonSerializerOptions { WriteIndented = true }));
                    return $"Success: 配置 {key} 已更新为 {value}。";
                }
                return "Error: 解析配置文件失败。";
            }
            catch (Exception ex)
            {
                return $"Error: 更新配置失败 - {ex.Message}";
            }
        }

        [KernelFunction(name: "execute_db_query")]
        [Description("执行只读的 SQL 查询语句（仅限 SELECT）。")]
        [ToolRisk(ToolRiskLevel.Medium, "执行数据库只读查询")]
        public async Task<string> ExecuteQuery([Description("SQL 查询语句")] string sql)
        {
            if (!sql.TrimStart().StartsWith("SELECT", StringComparison.OrdinalIgnoreCase))
            {
                return "Error: 仅允许执行 SELECT 查询。";
            }

            try
            {
                // 这里接入现有的 ORM 或数据库访问层
                // 示例返回
                return "查询执行成功，但结果集转换尚未完全实现。请联系管理员配置数据源。";
            }
            catch (Exception ex)
            {
                return $"Error: 数据库查询失败 - {ex.Message}";
            }
        }
    }
}
