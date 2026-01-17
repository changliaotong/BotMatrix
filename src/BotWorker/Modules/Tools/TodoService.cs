using System;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Common.Extensions; // For IsNull, IsNum, AsInt
using BotWorker.Domain.Entities;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.Tools
{
    public class TodoService : ITodoService
    {
        private readonly ITodoRepository _repository;
        private readonly ILogger<TodoService> _logger;

        public TodoService(ITodoRepository repository, ILogger<TodoService> logger)
        {
            _repository = repository;
            _logger = logger;
        }

        public const string format = "todo + 内容 [Dev/Test] [P1/P2/P3] 新增\ntodo - 数字 删除\ntodo #数字 [进度/done/P1/P2/P3/desc 内容] 更新\ntodo + #关键字 查询";
        public static string RetryMsg = "操作失败，请重试";

        public async Task<string> GetTodoResAsync(long groupId, string groupName, long qq, string name, string cmdName, string cmdPara)
        {
            _logger.LogInformation($"[Todo] GetTodoResAsync: cmdName='{{cmdName}}', cmdPara='{{cmdPara}}'", cmdName, cmdPara);
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

            if (cmdName == "+" || (cmdName.IsNull() && !cmdPara.IsNull() && !cmdPara.StartsWith('#')))
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
                    var existing = await _repository.GetByNoAsync(qq, todo_no);
                    if (existing == null)
                        return $"不存在#{todo_no}";

                    return await _repository.DeleteByNoAsync(qq, todo_no) == -1
                        ? RetryMsg
                        : $"成功删除#{todo_no}";
                }
                else if (cmdPara == "all")
                {
                     var all = await _repository.GetListAsync(qq);
                     int count = 0;
                     foreach(var t in all) {
                         await _repository.DeleteAsync(t);
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
                        var todo = await _repository.GetByNoAsync(qq, todo_no);
                        if (todo == null) return $"不存在#{todo_no}";

                        string updateVal = parts[1].ToUpper();

                        if (updateVal == "DONE")
                        {
                            todo.Progress = 100;
                            todo.Status = "Completed";
                            return await _repository.UpdateAsync(todo) ? $"待办#{todo_no}已完成！" : RetryMsg;
                        }
                        else if (updateVal == "DESC")
                        {
                            string desc = string.Join(" ", parts.Skip(2));
                            todo.Description = desc;
                            return await _repository.UpdateAsync(todo) ? $"待办#{todo_no}描述已更新" : RetryMsg;
                        }
                        else if (updateVal == "P1") 
                        { 
                            todo.Priority = "High"; 
                            return await _repository.UpdateAsync(todo) ? $"待办#{todo_no}优先级已更新为 High" : RetryMsg; 
                        }
                        else if (updateVal == "P2") 
                        { 
                            todo.Priority = "Medium"; 
                            return await _repository.UpdateAsync(todo) ? $"待办#{todo_no}优先级已更新为 Medium" : RetryMsg; 
                        }
                        else if (updateVal == "P3") 
                        { 
                            todo.Priority = "Low"; 
                            return await _repository.UpdateAsync(todo) ? $"待办#{todo_no}优先级已更新为 Low" : RetryMsg; 
                        }
                        else if (updateVal.IsNum())
                        {
                            int progress = updateVal.AsInt();
                            progress = Math.Clamp(progress, 0, 100);
                            string status = progress == 100 ? "Completed" : (progress > 0 ? "InProgress" : "Pending");
                            todo.Progress = progress;
                            todo.Status = status;
                            return await _repository.UpdateAsync(todo) ? $"待办#{todo_no}进度已更新为 {progress}%" : RetryMsg;
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

        private async Task<(int Result, int TodoNo)> AppendAsync(long groupId, long qq, string cmdPara, string category, string priority)
        {
            int todoNo = await _repository.GetMaxNoAsync(qq) + 1;
            var res = await _repository.InsertAsync(new Todo
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

        private async Task<string> GetTodoAsync(long qq, int todoNo)
        {
            var todo = await _repository.GetByNoAsync(qq, todoNo);
            if (todo == null) return "";
            return $"#{todo.TodoNo} [{todo.Category}] [{todo.Priority}] {todo.Progress}% {todo.Status} {todo.TodoTitle}\n描述: {todo.Description}";
        }

        private async Task<string> GetTodosAsync(long qq, string cmdPara, int topN = 5)
        {
            var list = await _repository.GetListAsync(qq);
            
            if (!string.IsNullOrEmpty(cmdPara))
            {
                if (cmdPara.Equals("dev", StringComparison.OrdinalIgnoreCase)) list = list.Where(t => t.Category == "Dev");
                else if (cmdPara.Equals("test", StringComparison.OrdinalIgnoreCase)) list = list.Where(t => t.Category == "Test");
                else if (cmdPara.Equals("p1", StringComparison.OrdinalIgnoreCase)) list = list.Where(t => t.Priority == "High");
                else list = list.Where(t => t.TodoTitle.Contains(cmdPara, StringComparison.OrdinalIgnoreCase));
            }

            var totalCount = list.Count();
            var displayList = list.Take(topN).ToList();

            if (!displayList.Any()) return $"太好了，没有todo";

            string result = "";
            result += $"{displayList.Count}/{totalCount}\n";
            
            foreach (var t in displayList)
            {
                string title = t.TodoTitle.Length > 20 ? t.TodoTitle.Substring(0, 20) : t.TodoTitle;
                result += $"#{t.TodoNo} [{t.Category}] [{t.Priority}] [{t.Progress}%] {title}\n";
            }

            return result;
        }
    }
}
