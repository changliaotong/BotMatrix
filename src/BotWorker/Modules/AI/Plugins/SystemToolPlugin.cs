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
        private readonly string _projectRoot;

        public SystemToolPlugin()
        {
            // 尝试获取项目根目录 (向上查找两级，假设在 src/BotWorker/bin/Debug/netX.0)
            var current = AppDomain.CurrentDomain.BaseDirectory;
            var dir = new DirectoryInfo(current);
            while (dir != null && !File.Exists(Path.Combine(dir.FullName, "BotMatrix.sln")))
            {
                dir = dir.Parent;
            }
            _projectRoot = dir?.FullName ?? AppDomain.CurrentDomain.BaseDirectory;
        }

        [KernelFunction(name: "read_code")]
        [Description("读取项目源代码文件的内容。")]
        [ToolRisk(ToolRiskLevel.Low, "读取系统源代码")]
        public async Task<string> ReadCode([Description("源代码文件相对路径 (从解决方案根目录开始)")] string filePath)
        {
            try
            {
                if (filePath.Contains("..")) return "Error: 禁止跨目录访问。";
                var fullPath = Path.Combine(_projectRoot, filePath);
                if (!File.Exists(fullPath)) return $"Error: 文件 {filePath} 不存在。";
                return await File.ReadAllTextAsync(fullPath);
            }
            catch (Exception ex)
            {
                return $"Error: 读取文件失败 - {ex.Message}";
            }
        }

        [KernelFunction(name: "write_code")]
        [Description("修改或创建项目源代码文件。")]
        [ToolRisk(ToolRiskLevel.High, "修改系统源代码")]
        public async Task<string> WriteCode(
            [Description("源代码文件相对路径")] string filePath,
            [Description("代码内容")] string content)
        {
            try
            {
                if (filePath.Contains("..")) return "Error: 禁止跨目录访问。";
                var fullPath = Path.Combine(_projectRoot, filePath);
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
                var configPath = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "appsettings.json");
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
