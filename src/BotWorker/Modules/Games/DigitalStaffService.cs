using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Models;
using BotWorker.Infrastructure.Persistence.ORM;
using BotWorker.Infrastructure.Utils.Schema;
using Microsoft.Extensions.Logging;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace BotWorker.Modules.Games
{
    public enum StaffRole
    {
        ProductManager, // 需求分析与规划
        Developer,      // 自动编程与系统升级
        Tester,         // 自动化测试与质量控制
        CustomerService,// 自动答疑与用户引导
        Sales,          // 自动营销与流量变现
        AfterSales      // 异常监测与系统维护
    }

    public class DigitalStaff : MetaDataGuid<DigitalStaff>
    {
        public override string TableName => "DigitalStaff";
        public override string KeyField => "Id";

        public string OwnerUserId { get; set; } = string.Empty;
        public string StaffName { get; set; } = string.Empty;
        public StaffRole Role { get; set; }
        public int Level { get; set; } = 1;
        public long TotalProfitGenerated { get; set; } = 0; // 累计创造收益
        public long SalaryToken { get; set; } = 0;          // 已消耗的虚拟薪资(Token)
        public long SalaryLimit { get; set; } = 1000000;    // 薪资上限
        public double KpiScore { get; set; } = 100.0;       // 平均绩效评分
        public string SystemPrompt { get; set; } = string.Empty; // 核心提示词
        public DateTime HireDate { get; set; } = DateTime.Now;
        public string CurrentStatus { get; set; } = "Idle"; // Idle, Working, Evolving
        public string AssignedTaskId { get; set; } = string.Empty; // 当前分配的任务ID
    }

    /// <summary>
    /// 认知记忆实体
    /// </summary>
    public class CognitiveMemory : MetaDataGuid<CognitiveMemory>
    {
        public override string TableName => "CognitiveMemories";
        public override string KeyField => "Id";

        public string StaffId { get; set; } = string.Empty; // 关联员工ID
        public string UserId { get; set; } = string.Empty;  // 关联用户ID (若为角色记忆则为空)
        public string Category { get; set; } = "General";   // 记忆类别
        public string Content { get; set; } = string.Empty; // 记忆内容
        public int Importance { get; set; } = 3;            // 重要程度 (1-5)
        public string Embedding { get; set; } = string.Empty; // 向量表示 (JSON)
        public DateTime LastSeen { get; set; } = DateTime.Now;
        public DateTime CreateTime { get; set; } = DateTime.Now;
    }

    /// <summary>
    /// 绩效考核记录
    /// </summary>
    public class StaffKpi : MetaDataGuid<StaffKpi>
    {
        public override string TableName => "StaffKpis";
        public override string KeyField => "Id";

        public string StaffId { get; set; } = string.Empty;
        public string MetricName { get; set; } = string.Empty; // 考核指标
        public double Score { get; set; } = 0;                 // 评分
        public string Detail { get; set; } = string.Empty;     // 详情/反馈
        public DateTime CreateTime { get; set; } = DateTime.Now;
    }

    /// <summary>
    /// 员工任务实体
    /// </summary>
    public class StaffTask : MetaDataGuid<StaffTask>
    {
        public override string TableName => "StaffTasks";
        public override string KeyField => "Id";

        public string Title { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
        public string TaskType { get; set; } = string.Empty; // Dev, Test, CS, Sales
        public string Status { get; set; } = "Pending"; // Pending, InProgress, Completed, Failed
        public string CreatorUserId { get; set; } = string.Empty;
        public string ExecutorStaffId { get; set; } = string.Empty;
        public string Result { get; set; } = string.Empty;
        public DateTime CreateTime { get; set; } = DateTime.Now;
        public DateTime? CompleteTime { get; set; }
    }

    [BotPlugin(
        Id = "core.digital_staff",
        Name = "数字员工管理系统",
        Version = "1.0.0",
        Author = "BotMatrix Cyber",
        Description = "管理您的数字劳动力，包括自动编程、需求分析与自动销售员工。",
        Category = "Core"
    )]
    public class DigitalStaffService : IPlugin
    {
        private readonly ILogger<DigitalStaffService>? _logger;
        private IRobot? _robot;

        public DigitalStaffService() { }
        public DigitalStaffService(ILogger<DigitalStaffService> logger)
        {
            _logger = logger;
        }

        public List<Intent> Intents => [
            new() { Name = "雇佣员工", Keywords = ["雇佣", "招聘", "staff"] },
            new() { Name = "员工列表", Keywords = ["我的员工", "团队"] },
            new() { Name = "指派任务", Keywords = ["指派", "开发", "销售"] }
        ];

        public async Task InitAsync(IRobot robot)
        {
            _robot = robot;
            await EnsureTablesCreatedAsync();

            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "赛博人才中心",
                Commands = ["雇佣", "招聘", "我的员工", "团队", "派单", "公司"],
                Description = "【公司】查看看板；【雇佣 角色 名字】招聘员工；【派单 标题 描述】发布真实任务"
            }, HandleCommandAsync);

            // 启动后台任务处理器
            _ = Task.Run(async () => {
                while (true)
                {
                    try { await ProcessStaffTasksAsync(); }
                    catch (Exception ex) { _logger?.LogError(ex, "任务处理器异常"); }
                    await Task.Delay(TimeSpan.FromMinutes(1));
                }
            });
        }

        public async Task StopAsync()
        {
            _logger?.LogInformation("数字员工服务已停止");
            await Task.CompletedTask;
        }

        private async Task EnsureTablesCreatedAsync()
        {
            try
            {
                var checkStaff = await DigitalStaff.QueryScalarAsync<int>($"SELECT COUNT(*) FROM {DigitalStaff.DbName}.INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'DigitalStaff'");
                if (checkStaff == 0)
                {
                    await DigitalStaff.ExecAsync(SchemaSynchronizer.GenerateCreateTableSql<DigitalStaff>());
                }

                var checkTask = await StaffTask.QueryScalarAsync<int>($"SELECT COUNT(*) FROM {StaffTask.DbName}.INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'StaffTasks'");
                if (checkTask == 0)
                {
                    await StaffTask.ExecAsync(SchemaSynchronizer.GenerateCreateTableSql<StaffTask>());
                }

                var checkMemory = await CognitiveMemory.QueryScalarAsync<int>($"SELECT COUNT(*) FROM {CognitiveMemory.DbName}.INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'CognitiveMemories'");
                if (checkMemory == 0)
                {
                    await CognitiveMemory.ExecAsync(SchemaSynchronizer.GenerateCreateTableSql<CognitiveMemory>());
                }

                var checkKpi = await StaffKpi.QueryScalarAsync<int>($"SELECT COUNT(*) FROM {StaffKpi.DbName}.INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'StaffKpis'");
                if (checkKpi == 0)
                {
                    await StaffKpi.ExecAsync(SchemaSynchronizer.GenerateCreateTableSql<StaffKpi>());
                }
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "DigitalStaffService 数据库初始化失败");
                throw;
            }
        }

        private async Task<string> HireStaffAsync(IPluginContext ctx, string[] args)
        {
            if (args.Length < 2) return "❌ 请输入：雇佣 [名字] [职位:PM/Dev/Sales/AfterSales]";

            var name = args[0];
            var roleStr = args[1].ToLower();
            StaffRole role = roleStr switch
            {
                "pm" or "产品" or "产品经理" => StaffRole.ProductManager,
                "dev" or "代码" or "开发" or "工程师" => StaffRole.Developer,
                "sales" or "销售" or "市场" => StaffRole.Sales,
                "aftersales" or "客服" or "售后" => StaffRole.AfterSales,
                _ => StaffRole.Developer
            };

            var staff = new DigitalStaff
            {
                StaffName = name,
                Role = role,
                OwnerUserId = ctx.UserId,
                SystemPrompt = $"你是一个专业的{role}。请高效完成分配给你的任务。",
                Level = 1,
                KpiScore = 100.0,
                CurrentStatus = "Idle"
            };

            await staff.InsertAsync();
            return $"🎉 恭喜！您已成功雇佣【{name}】（职位：{role}）。现在可以尝试【派单】了。";
        }

        private async Task<string> HandleCommandAsync(IPluginContext ctx, string[] args)
        {
            if (string.IsNullOrWhiteSpace(ctx.RawMessage)) return string.Empty;
            
            var cmd = ctx.RawMessage.Trim().Split(' ')[0].TrimStart('!', '！', '/', ' ');

            return cmd switch
            {
                "雇佣" or "招聘" => await HireStaffAsync(ctx, args),
                "我的员工" or "团队" or "公司" => await GetCompanyBoardAsync(ctx.UserId),
                "派单" or "发布任务" => await CreateTaskAsync(ctx, args),
                "认领" => await ClaimTaskAsync(ctx, args),
                _ => "💡 赛博人才中心：使用【公司】查看看板，【雇佣】招聘人才，【派单】发布真实工作。"
            };
        }

        private async Task<string> CreateTaskAsync(IPluginContext ctx, string[] args)
        {
            if (args.Length < 2) return "❌ 格式错误：请使用【派单 标题 描述】。";

            var task = new StaffTask
            {
                Title = args[0],
                Description = args[1],
                CreatorUserId = ctx.UserId,
                TaskType = "Dev",
                Status = "Pending"
            };

            await task.InsertAsync();
            return $"✅ 任务【{task.Title}】已发布。空闲员工将自动尝试【认领】。";
        }

        private async Task<string> ClaimTaskAsync(IPluginContext ctx, string[] args)
        {
            var freeStaff = await DigitalStaff.QueryWhere("OwnerUserId = @p1 AND CurrentStatus = 'Idle'", DigitalStaff.SqlParams(("@p1", ctx.UserId)));
            var staff = freeStaff.FirstOrDefault();

            if (staff == null) return "❌ 您当前没有空闲的员工。";

            var pendingTasks = await StaffTask.QueryWhere("Status = 'Pending' ORDER BY CreateTime ASC");
            var task = pendingTasks.FirstOrDefault();

            if (task == null) return "📭 任务池目前是空的。";

            staff.CurrentStatus = "Working";
            staff.AssignedTaskId = task.Guid.ToString();
            await staff.UpdateAsync();

            task.Status = "InProgress";
            task.ExecutorStaffId = staff.Guid.ToString();
            await task.UpdateAsync();

            return $"💼 员工【{staff.StaffName}】已认领任务：{task.Title}，开始投入生产。";
        }

        private async Task ProcessStaffTasksAsync()
        {
            var workingStaff = await DigitalStaff.QueryWhere("CurrentStatus = 'Working'");
            foreach (var staff in workingStaff)
            {
                if (!Guid.TryParse(staff.AssignedTaskId, out var taskGuid)) continue;
                var task = await StaffTask.LoadAsync(taskGuid);
                if (task == null || task.Status != "InProgress") continue;

                // 模拟工作进度与真实产出
                switch (staff.Role)
                {
                    case StaffRole.Developer:
                        await ExecuteDevTaskAsync(staff, task);
                        break;
                    case StaffRole.Sales:
                        await ExecuteSalesTaskAsync(staff, task);
                        break;
                    case StaffRole.CustomerService:
                        await ExecuteCSTaskAsync(staff, task);
                        break;
                    case StaffRole.AfterSales:
                        await ExecuteAfterSalesTaskAsync(staff, task);
                        break;
                }
            }
        }

        private async Task ExecuteCSTaskAsync(DigitalStaff staff, StaffTask task)
        {
            // CustomerService 逻辑：代为询问 MatrixOracle
            if (_robot != null)
            {
                var oracleResponse = await _robot.CallSkillAsync("oracle.query", null!, new[] { task.Description });
                
                task.Status = "Completed";
                task.Result = oracleResponse?.ToString() ?? "先知暂未回应。";
                task.CompleteTime = DateTime.Now;
                await task.UpdateAsync();

                staff.CurrentStatus = "Idle";
                staff.AssignedTaskId = string.Empty;
                await staff.UpdateAsync();

                await _robot.SendMessageAsync("system", "bot", null, staff.OwnerUserId, 
                    $"🎧 客服员工【{staff.StaffName}】已为您获取答案：\n\n{task.Result}");
            }
        }

        private async Task ExecuteAfterSalesTaskAsync(DigitalStaff staff, StaffTask task)
        {
            // AfterSales 逻辑：系统维护
            task.Status = "Completed";
            task.Result = "已完成系统例行检查，清理了冗余的临时数据。";
            task.CompleteTime = DateTime.Now;
            await task.UpdateAsync();

            staff.CurrentStatus = "Idle";
            staff.AssignedTaskId = string.Empty;
            await staff.UpdateAsync();

            if (_robot != null)
            {
                await _robot.SendMessageAsync("system", "bot", null, staff.OwnerUserId,
                    $"🔧 售后员工【{staff.StaffName}】报告：{task.Result}");
            }
        }

        /// <summary>
        /// AI 审计员：自动评估任务产出并打分 (迁移自 Go: AuditTask)
        /// </summary>
        private async Task AuditTaskAsync(DigitalStaff staff, StaffTask task)
        {
            if (_robot?.AI == null) return;

            string auditPrompt = $@"你是一个严苛的首席执行官。请对以下员工完成的任务进行绩效评估。
员工：{staff.StaffName} (职位: {staff.Role})
任务标题：{task.Title}
任务描述：{task.Description}
执行结果：
""""""
{task.Result}
""""""

请根据执行结果给出：
1. 评分 (0-100)；
2. 简短的改进建议。
输出格式：[SCORE:分数] 建议内容";

            string auditResult = await _robot.AI.ChatAsync(auditPrompt);
            var match = System.Text.RegularExpressions.Regex.Match(auditResult, @"\[SCORE:(\d+)\]\s*(.*)");
            
            double score = 80.0;
            string detail = auditResult;

            if (match.Success)
            {
                double.TryParse(match.Groups[1].Value, out score);
                detail = match.Groups[2].Value;
            }

            await RecordKpiAsync(staff.Guid.ToString(), "ai_audit", score, detail);
        }

        private async Task ExecuteDevTaskAsync(DigitalStaff staff, StaffTask task)
        {
            if (_robot?.AI == null) return;

            _logger?.LogInformation($"员工 {staff.StaffName} 开始执行开发任务: {task.Title}");

            // 获取可用技能描述
            var skillsDescription = string.Join("\n", _robot.Skills.Select(s => $"- {s.Capability.Name}: {s.Capability.Description} (指令: {string.Join("/", s.Capability.Commands)})"));

            string prompt = $@"你是一个高级全栈工程师。
任务标题：{task.Title}
任务描述：{task.Description}

当前可用的系统技能（Skills）：
{skillsDescription}

请根据任务描述执行以下逻辑：
1. 分析是否需要调用上述技能来辅助完成任务。
2. 如果需要调用技能，请在回复的最开始输出：[CALL_SKILL:技能名称:参数1,参数2...]
3. 给出详细的技术实现方案或代码。
要求：分析需求、给出核心逻辑代码块、说明注意事项。";

            string result = await _robot.AI.ChatAsync(prompt);

            // 处理 AI 的技能调用意图
            if (result.StartsWith("[CALL_SKILL:"))
            {
                var match = System.Text.RegularExpressions.Regex.Match(result, @"\[CALL_SKILL:(.*?):(.*?)\]");
                if (match.Success)
                {
                    var skillName = match.Groups[1].Value;
                    var args = match.Groups[2].Value.Split(',');
                    _logger?.LogInformation($"员工 {staff.StaffName} 决定调用技能: {skillName}");
                    
                    var skillResult = await _robot.CallSkillAsync(skillName, null!, args);
                    result = result.Substring(match.Length).Trim();
                    result = $"[技能调用成果: {skillName}]\n{skillResult}\n\n[后续分析]\n{result}";
                }
            }

            task.Status = "Completed";
            task.Result = result;
            task.CompleteTime = DateTime.Now;
            await task.UpdateAsync();

            staff.CurrentStatus = "Idle";
            staff.AssignedTaskId = string.Empty;
            await staff.UpdateAsync();

            // AI 自动审计绩效
            await AuditTaskAsync(staff, task);

            await _robot.SendMessageAsync("system", "bot", null, staff.OwnerUserId,
                $"💻 开发员工【{staff.StaffName}】已完成任务【{task.Title}】：\n\n{task.Result}");
        }

        private async Task RecordKpiAsync(string staffId, string metric, double score, string detail)
        {
            var kpi = new StaffKpi
            {
                StaffId = staffId,
                MetricName = metric,
                Score = score,
                Detail = detail,
                CreateTime = DateTime.Now
            };
            await kpi.InsertAsync();

            // 更新员工平均分
            var staff = await DigitalStaff.LoadAsync(new Guid(staffId));
            if (staff != null)
            { 
                var kpis = await StaffKpi.QueryWhere("StaffId = @p1", StaffKpi.SqlParams(("@p1", staffId)));
                staff.KpiScore = kpis.Average(k => k.Score);
                await staff.UpdateAsync();

                // 检查是否触发自动进化
                if (staff.KpiScore > 95.0 && kpis.Count() % 5 == 0)
                {
                    _ = AutoEvolveAsync(staff);
                }
            }
        }

        /// <summary>
        /// 自动进化逻辑 (迁移自 Go: AutoEvolve)
        /// </summary>
        private async Task AutoEvolveAsync(DigitalStaff staff)
        {
            if (staff.CurrentStatus == "Evolving" || _robot?.AI == null) return;

            staff.CurrentStatus = "Evolving";
            await staff.UpdateAsync();

            try
            {
                var kpis = await StaffKpi.QueryListAsync(new QueryOptions 
                { 
                    FilterSql = "StaffId = @p1", 
                    OrderBy = "CreateTime DESC", 
                    Top = 10, 
                    Parameters = StaffKpi.SqlParams(("@p1", staff.Guid.ToString())) 
                });
                if (!kpis.Any())
                {
                    staff.CurrentStatus = "Idle";
                    await staff.UpdateAsync();
                    return;
                }

                string feedback = string.Join("\n", kpis.Where(k => !string.IsNullOrEmpty(k.Detail)).Select(k => $"- [{k.CreateTime:yyyy-MM-dd}] {k.MetricName}: {k.Detail}"));
                
                string systemPromptTemplate = @"你是一个资深的 AI 提示词架构师。你的任务是根据数字员工的当前系统提示词和最近的 KPI 绩效反馈，优化其提示词。
数字员工信息：
- 姓名：{0}
- 职位：{1}

当前系统提示词：
""""""
{2}
""""""

最近的绩效反馈与评分（平均分：{3:F2}）：
""""""
{4}
""""""

请分析反馈中的不足（如：专业度不够、回复太慢、语气生硬、未遵循规范等），并输出一个优化后的、更强大的系统提示词。
要求：
1. 保持原有的人设特征。
2. 针对性地解决反馈中提到的问题。
3. 增强对复杂场景的处理能力。
4. 只输出优化后的系统提示词内容，不要包含其他解释。";

                string finalPrompt = string.Format(systemPromptTemplate, staff.StaffName, staff.Role, staff.SystemPrompt, staff.KpiScore, feedback);

                string newPrompt = await _robot.AI.ChatAsync(finalPrompt);
                
                if (!string.IsNullOrEmpty(newPrompt) && newPrompt != staff.SystemPrompt)
                {
                    staff.SystemPrompt = newPrompt;
                    _logger?.LogInformation($"员工 {staff.StaffName} 提示词已自动优化。");
                }

                staff.CurrentStatus = "Idle";
                await staff.UpdateAsync();

                // 记录进化记录
                await RecordKpiAsync(staff.Guid.ToString(), "auto_evolution", staff.KpiScore, $"提示词已自动优化。旧评分: {staff.KpiScore:F2}。反馈摘要: {kpis.Count()} 条记录已处理。");
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, $"员工 {staff.StaffName} 进化失败");
                staff.CurrentStatus = "Idle";
                await staff.UpdateAsync();
            }
        }

        /// <summary>
        /// 记忆固化逻辑 (迁移自 Go: ConsolidateMemories)
        /// </summary>
        private async Task ConsolidateMemoriesAsync(string staffId)
        {
            if (_robot?.AI == null) return;

            var memories = await CognitiveMemory.QueryWhere("StaffId = @p1 ORDER BY Category, CreateTime ASC", CognitiveMemory.SqlParams(("@p1", staffId)));
            if (memories.Count() < 10) return;

            string prompt = "你是一个记忆管理专家。以下是碎片化记忆片段。请将这些记忆进行逻辑合并、去重并提炼。\n" +
                           "规则：\n1. 合并相似内容；2. 保持分类清晰；3. 提炼出更有深度的洞察；4. 格式：[类别] 内容。\n\n记忆片段：\n";
            prompt += string.Join("\n", memories.Select(m => $"- [{m.Category}] {m.Content}"));

            string consolidated = await _robot.AI.ChatAsync(prompt);
            if (string.IsNullOrWhiteSpace(consolidated)) return;

            var lines = consolidated.Split('\n', StringSplitOptions.RemoveEmptyEntries);
            var newMemories = new List<CognitiveMemory>();
            foreach (var line in lines)
            {
                var match = System.Text.RegularExpressions.Regex.Match(line, @"^\[(.*?)\]\s*(.*)$");
                if (match.Success)
                {
                    newMemories.Add(new CognitiveMemory
                    {
                        StaffId = staffId,
                        Category = match.Groups[1].Value,
                        Content = match.Groups[2].Value,
                        Importance = 3,
                        CreateTime = DateTime.Now,
                        LastSeen = DateTime.Now
                    });
                }
            }

            if (newMemories.Any())
            {
                // 使用事务替换记忆
                var sqls = new List<string> { $"DELETE FROM CognitiveMemories WHERE StaffId = '{staffId}'" };
                foreach (var m in newMemories)
                {
                    sqls.Add($"INSERT INTO CognitiveMemories (Id, StaffId, Category, Content, Importance, CreateTime, LastSeen) " +
                             $"VALUES ('{Guid.NewGuid()}', '{staffId}', '{m.Category}', '{m.Content}', {m.Importance}, '{m.CreateTime:yyyy-MM-dd HH:mm:ss}', '{m.LastSeen:yyyy-MM-dd HH:mm:ss}')");
                }
                BotWorker.Infrastructure.Persistence.Database.SQLConn.ExecTrans(sqls.ToArray());
                _logger?.LogInformation($"员工 {staffId} 记忆提炼完成，新增 {newMemories.Count} 条记忆。");
            }
        }

        private async Task ExecuteSalesTaskAsync(DigitalStaff staff, StaffTask task)
        {
            if (_robot?.AI == null) return;

            _logger?.LogInformation($"员工 {staff.StaffName} 开始执行销售/营销任务: {task.Title}");

            string prompt = $@"你是一个天才营销专家。
任务：{task.Title}
背景：{task.Description}

请生成一段极具吸引力的文案，用于推广此产品或服务。
要求：
1. 吸引眼球的标题；
2. 痛点分析与解决方案；
3. 强力行动号召 (CTA)。";

            string result = await _robot.AI.ChatAsync(prompt);

            task.Status = "Completed";
            task.Result = result;
            task.CompleteTime = DateTime.Now;
            await task.UpdateAsync();

            staff.CurrentStatus = "Idle";
            staff.AssignedTaskId = string.Empty;
            await staff.UpdateAsync();

            // 销售任务可能会产生虚拟收益
            staff.TotalProfitGenerated += 500; 
            await staff.UpdateAsync();

            // AI 自动审计绩效
            await AuditTaskAsync(staff, task);

            await _robot.SendMessageAsync("system", "bot", null, staff.OwnerUserId,
                $"💰 销售员工【{staff.StaffName}】已完成任务【{task.Title}】，预计带来收益 500 Credits：\n\n{task.Result}");
        }

        private async Task<string> GetCompanyBoardAsync(string userId)
        {
            var staffs = await DigitalStaff.QueryWhere("OwnerUserId = @p1", DigitalStaff.SqlParams(("@p1", userId)));
            if (!staffs.Any()) return "🏢 您目前还没有组建团队。使用【雇佣】来开始运营吧！";

            var sb = new System.Text.StringBuilder();
            sb.AppendLine("┏━━━━━━ 赛博公司看板 ━━━━━━┓");
            foreach (var s in staffs)
            {
                string icon = s.Role switch { StaffRole.ProductManager => "📝", StaffRole.Developer => "💻", StaffRole.Sales => "📈", _ => "👤" };
                string status = s.CurrentStatus == "Working" ? "⚙️ 生产中" : "☕ 待命";
                sb.AppendLine($"┃ {icon} {s.StaffName.PadRight(10)} | Lv.{s.Level} | {status}");
            }
            sb.AppendLine("┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫");
            var pending = await StaffTask.QueryAsync("WHERE Status = 'Pending'");
            sb.AppendLine($"┃ � 待处理任务: {pending.Count()} 个");
            sb.AppendLine("┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛");
            return sb.ToString();
        }
    }
}
