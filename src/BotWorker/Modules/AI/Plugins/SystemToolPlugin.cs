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
