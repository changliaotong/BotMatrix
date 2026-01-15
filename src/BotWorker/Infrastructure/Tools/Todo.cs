using System;
using System.Threading.Tasks;
using System.Linq;
using System.Collections.Generic;
using BotWorker.Domain.Repositories;
using BotWorker.Domain.Models.BotMessages;
using Microsoft.Extensions.DependencyInjection;
using Dapper.Contrib.Extensions;
using BotWorker.Infrastructure.Utils.Schema.Attributes;
using BotWorker.Infrastructure.Extensions; // For IsNull, IsNum, AsInt

namespace BotWorker.Infrastructure.Tools
{
    [Table("Todo")]
    public class Todo
    {
        private static ITodoRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<ITodoRepository>() 
            ?? throw new InvalidOperationException("ITodoRepository not registered");

        [Key]
        public long Id { get; set; }
        public long GroupId { get; set; }
        public long UserId { get; set; }
        public int TodoNo { get; set; }
        public string TodoTitle { get; set; } = "";
        public string Description { get; set; } = ""; // 详细描述
        public string Priority { get; set; } = "Medium"; // Low, Medium, High
        public int Progress { get; set; } = 0; // 进度 0-100
        public string Status { get; set; } = "Pending"; // Pending, InProgress, Completed
        public string Category { get; set; } = "Todo"; // Todo, Dev, Test
        public DateTime? DueDate { get; set; } // 截止日期
        public DateTime InsertDate { get; set; } = DateTime.Now;

        public static async Task EnsureTableCreatedAsync()
        {
            // Removed legacy SQL migration
            await Task.CompletedTask;
        }

        public const string format = "todo + 内容 [Dev/Test] [P1/P2/P3] 新增\ntodo - 数字 删除\ntodo #数字 [进度/done/P1/P2/P3/desc 内容] 更新\ntodo + #关键字 查询";
        public static string RetryMsg = "操作失败，请重试";

        // todo
        public static async Task<string> GetTodoResAsync(long groupId, string groupName, long qq, string name, string cmdName, string cmdPara)
        {
            Console.WriteLine($"[Todo] GetTodoResAsync: cmdName='{cmdName}', cmdPara='{cmdPara}'");
            cmdName = cmdName.ToLower();
            if (cmdName == "todo" || cmdName == "td")
            {
                if (cmdPara.StartsWith("+"))
                {
                    cmdName = "+";
                    cmdPara = cmdPara[1..].Trim();
                }
                else if (cmdPara.StartsWith("-"))
                {
                    cmdName = "-";
                    cmdPara = cmdPara[1..].Trim();
                }
                else if (cmdPara.StartsWith("#"))
                {
                    // 保持 cmdName 为 todo，后面会处理 # 开头的 cmdPara
                }
                else if (cmdPara.IsNull())
                {
                    // 列表查询
                }
                else
                {
                    // 默认当作搜索
                    return await GetTodosAsync(qq, cmdPara);
                }
            }
            Console.WriteLine($"[Todo] Processed: cmdName='{cmdName}', cmdPara='{cmdPara}'");

            if (cmdName == "+" || cmdName.IsNull() && !cmdPara.IsNull() && !cmdPara.StartsWith('#'))
            {
                // 支持 todo + 内容 Dev P1 或 todo + 内容 Test
                string category = "Todo";
                string priority = "Medium";

                if (cmdPara.Contains(" Dev", StringComparison.OrdinalIgnoreCase))
                {
                    category = "Dev";
                    cmdPara = cmdPara.Replace(" Dev", "", StringComparison.OrdinalIgnoreCase).Trim();
                }
                else if (cmdPara.Contains(" Test", StringComparison.OrdinalIgnoreCase))
                {
                    category = "Test";
                    cmdPara = cmdPara.Replace(" Test", "", StringComparison.OrdinalIgnoreCase).Trim();
                }

                if (cmdPara.Contains(" P1", StringComparison.OrdinalIgnoreCase)) { priority = "High"; cmdPara = cmdPara.Replace(" P1", "", StringComparison.OrdinalIgnoreCase).Trim(); }
                else if (cmdPara.Contains(" P2", StringComparison.OrdinalIgnoreCase)) { priority = "Medium"; cmdPara = cmdPara.Replace(" P2", "", StringComparison.OrdinalIgnoreCase).Trim(); }
                else if (cmdPara.Contains(" P3", StringComparison.OrdinalIgnoreCase)) { priority = "Low"; cmdPara = cmdPara.Replace(" P3", "", StringComparison.OrdinalIgnoreCase).Trim(); }

                var res = await AppendAsync(groupId, qq, cmdPara, category, priority);
                return res.Result == -1
                    ? RetryMsg
                    : $"待办添加成功，#{res.TodoNo} [{category}] [{priority}]";
            }
            else if (cmdName == "-" || cmdName == "删除")
            {
                if (cmdPara.StartsWith("#")) cmdPara = cmdPara[1..];
                if (cmdPara.IsNum())
                {
                    int todo_no = cmdPara.AsInt();
                    var existing = await Repository.GetByNoAsync(qq, todo_no);
                    if (existing == null)
                        return $"不存在#{todo_no}";

                    return await Repository.DeleteByNoAsync(qq, todo_no) == -1
                        ? RetryMsg
                        : $"成功删除#{todo_no}";
                }
                else if (cmdPara == "all")
                {
                     // Delete all for user
                     // I need to add DeleteAllAsync to repository or just execute sql
                     // For now, I'll use simple implementation
                     var all = await Repository.GetListAsync(qq);
                     int count = 0;
                     foreach(var t in all) {
                         await Repository.DeleteAsync(t);
                         count++;
                     }
                     return $"真棒！全部完成！";
                }
                else
                    return "删除参数只能是数字";
            }
            else if (cmdPara.StartsWith('#'))
            {
                var parts = cmdPara.Split(' ', StringSplitOptions.RemoveEmptyEntries);
                if (parts.Length >= 2 && parts[0].StartsWith('#'))
                {
                    // 更新进度或属性: todo #1 50 或 todo #1 done 或 todo #1 P1
                    string todoNoStr = parts[0][1..];
                    if (todoNoStr.IsNum())
                    {
                        int todo_no = todoNoStr.AsInt();
                        var todo = await Repository.GetByNoAsync(qq, todo_no);
                        if (todo == null) return $"不存在#{todo_no}";

                        string updateVal = parts[1].ToUpper();

                        if (updateVal == "DONE")
                        {
                            todo.Progress = 100;
                            todo.Status = "Completed";
                            return await Repository.UpdateAsync(todo) ? $"待办#{todo_no}已完成！" : RetryMsg;
                        }
                        else if (updateVal == "DESC")
                        {
                            string desc = string.Join(" ", parts.Skip(2));
                            todo.Description = desc;
                            return await Repository.UpdateAsync(todo) ? $"待办#{todo_no}描述已更新" : RetryMsg;
                        }
                        else if (updateVal == "P1") 
                        { 
                            todo.Priority = "High"; 
                            return await Repository.UpdateAsync(todo) ? $"待办#{todo_no}优先级已更新为 High" : RetryMsg; 
                        }
                        else if (updateVal == "P2") 
                        { 
                            todo.Priority = "Medium"; 
                            return await Repository.UpdateAsync(todo) ? $"待办#{todo_no}优先级已更新为 Medium" : RetryMsg; 
                        }
                        else if (updateVal == "P3") 
                        { 
                            todo.Priority = "Low"; 
                            return await Repository.UpdateAsync(todo) ? $"待办#{todo_no}优先级已更新为 Low" : RetryMsg; 
                        }
                        else if (updateVal.IsNum())
                        {
                            int progress = updateVal.AsInt();
                            progress = Math.Clamp(progress, 0, 100);
                            string status = progress == 100 ? "Completed" : (progress > 0 ? "InProgress" : "Pending");
                            todo.Progress = progress;
                            todo.Status = status;
                            return await Repository.UpdateAsync(todo) ? $"待办#{todo_no}进度已更新为 {progress}%" : RetryMsg;
                        }
                    }
                }
                else
                {
                    // 查询特定ID: todo #1
                    string todoNoStr = cmdPara[1..];
                    if (todoNoStr.IsNum())
                    {
                        string res = await GetTodoAsync(qq, todoNoStr.AsInt());
                        return res == "" ? $"没有#{todoNoStr}" : res;
                    }
                }
            }
            return await GetTodosAsync(qq, cmdPara);
        }

        public static string GetTodoRes(long groupId, string groupName, long qq, string name, string cmdName, string cmdPara)
            => GetTodoResAsync(groupId, groupName, qq, name, cmdName, cmdPara).GetAwaiter().GetResult();

        // 新增todo
        public static async Task<(int Result, int TodoNo)> AppendAsync(long groupId, long qq, string cmdPara, string category, string priority)
        {
            int todoNo = await Repository.GetMaxNoAsync(qq) + 1;
            var res = await Repository.InsertAsync(new Todo
            {
                GroupId = groupId,
                UserId = qq,
                TodoNo = todoNo,
                TodoTitle = cmdPara,
                Category = category,
                Priority = priority,
                Progress = 0,
                Status = "Pending"
            });
            return (res, todoNo);
        }

        public static int Append(long groupId, long qq, string cmdPara, string category, string priority, out int todoNo)
        {
            var res = AppendAsync(groupId, qq, cmdPara, category, priority).GetAwaiter().GetResult();
            todoNo = res.TodoNo;
            return res.Result;
        }

        // todo 数量
        public static async Task<long> CountAsync(long qq)
        {
            // Use GetListAsync to count (a bit inefficient but consistent with interface)
            var list = await Repository.GetListAsync(qq);
            return list.Count();
        }

        public static long Count(long qq)
            => CountAsync(qq).GetAwaiter().GetResult();

        // 得到 todoNo 的 todo
        public static async Task<string> GetTodoAsync(long qq, int todoNo)
        {
            var todo = await Repository.GetByNoAsync(qq, todoNo);
            if (todo == null) return "";
            // "#{0} [{1}] [{2}] {3}% {4} {5}\n描述: {6}"
            // TodoNo, Category, Priority, Progress, Status, TodoTitle, Description
            return $"#{todo.TodoNo} [{todo.Category}] [{todo.Priority}] {todo.Progress}% {todo.Status} {todo.TodoTitle}\n描述: {todo.Description}";
        }

        public static string GetTodo(long qq, int todoNo)
            => GetTodoAsync(qq, todoNo).GetAwaiter().GetResult();

        // todo 列表
        public static async Task<string> GetTodosAsync(long qq, string cmdPara, int topN = 5)
        {
            // Filter is handled in Repository.GetListAsync partially (keyword)
            // But repo method signature is GetListAsync(userId, keyword)
            // Existing logic handles 'dev', 'test', 'p1' specially.
            // I should update Repository to handle these or filter in memory.
            // Since dataset is small per user, memory filter is fine.
            
            var list = await Repository.GetListAsync(qq); // Get all for user
            
            if (!string.IsNullOrEmpty(cmdPara))
            {
                if (cmdPara.Equals("dev", StringComparison.OrdinalIgnoreCase)) list = list.Where(t => t.Category == "Dev");
                else if (cmdPara.Equals("test", StringComparison.OrdinalIgnoreCase)) list = list.Where(t => t.Category == "Test");
                else if (cmdPara.Equals("p1", StringComparison.OrdinalIgnoreCase)) list = list.Where(t => t.Priority == "High");
                else list = list.Where(t => t.TodoTitle.Contains(cmdPara, StringComparison.OrdinalIgnoreCase));
            }

            var totalCount = await CountAsync(qq);
            var displayList = list.Take(topN).ToList();

            if (!displayList.Any()) return $"太好了，没有todo";

            string result = "";
            // Header: {c}/{total}
            result += $"{displayList.Count}/{totalCount}\n";
            
            foreach (var t in displayList)
            {
                // "#{0} [{1}] [{2}] [{3}%] {4}\n"
                // TodoNo, Category, Priority, Progress, TodoTitle(truncated 20)
                string title = t.TodoTitle.Length > 20 ? t.TodoTitle.Substring(0, 20) : t.TodoTitle;
                result += $"#{t.TodoNo} [{t.Category}] [{t.Priority}] [{t.Progress}%] {title}\n";
            }

            return result;
        }

        public static string GetTodos(long qq, string cmdPara, int topN = 5)
            => GetTodosAsync(qq, cmdPara, topN).GetAwaiter().GetResult();
    }
}
