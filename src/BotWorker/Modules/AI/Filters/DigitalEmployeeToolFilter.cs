using Microsoft.SemanticKernel;
using System.Threading.Tasks;
using System.Reflection;
using System.Text.Json;
using BotWorker.Modules.AI.Tools;
using System.Linq;

namespace BotWorker.Modules.AI.Filters
{
    /// <summary>
    /// 遵循“数字员工工具接口规范 v1”的函数调用过滤器
    /// </summary>
    public class DigitalEmployeeToolFilter : IFunctionInvocationFilter
    {
        private readonly IToolAuditService _auditService;
        private readonly string _taskId;
        private readonly string _staffId;

        public DigitalEmployeeToolFilter(IToolAuditService auditService, string taskId = "system", string staffId = "system")
        {
            _auditService = auditService;
            _taskId = taskId;
            _staffId = staffId;
        }

        public async Task OnFunctionInvocationAsync(FunctionInvocationContext context, Func<FunctionInvocationContext, Task> next)
        {
            // 1. 获取工具元数据和风险等级
            var methodInfo = (context.Function as KernelFunction)?.Metadata; // 这种方式拿不到 Attribute
            
            // 尝试通过反射获取属性 (SK 的 KernelFunction 内部通常包装了真实的 MethodInfo)
            var riskAttr = GetRiskAttribute(context.Function);
            var riskLevel = riskAttr?.RiskLevel ?? ToolRiskLevel.Low;
            var toolName = $"{context.Function.PluginName}.{context.Function.Name}";
            
            var inputJson = JsonSerializer.Serialize(context.Arguments);

            // 2. 审计记录：开始调用
            var logId = await _auditService.LogCallAsync(_taskId, _staffId, toolName, inputJson, riskLevel);

            // 3. 风险控制：High 风险拦截 (规范 4.2: 必须人工确认)
            if (riskLevel == ToolRiskLevel.High)
            {
                await _auditService.MarkAsPendingApprovalAsync(logId);
                context.Result = new FunctionResult(context.Function, "ERROR: 该操作属于高风险行为，必须经过人工审批。请联系管理员或在审批面板通过后重试。");
                return;
            }

            try
            {
                // 4. 执行 Executor 职责
                await next(context);

                // 5.审计记录：执行结果
                var outputJson = JsonSerializer.Serialize(context.Result?.GetValue<object>());
                await _auditService.UpdateResultAsync(logId, outputJson, "Success");
            }
            catch (Exception ex)
            {
                // 审计记录：失败
                await _auditService.UpdateResultAsync(logId, ex.Message, "Failed");
                throw;
            }
        }

        private ToolRiskAttribute? GetRiskAttribute(KernelFunction function)
        {
            // 由于 SK 隐藏了原始 MethodInfo，我们可能需要通过一些技巧获取
            // 或者在创建插件时显式把风险等级存入 Metadata
            if (function.Metadata.Parameters != null)
            {
                // 检查是否有我们注入的元数据
                // 另一种方式是利用反射，SK 的某些实现类暴露了 MethodInfo
                var prop = function.GetType().GetProperty("MethodInfo", BindingFlags.Instance | BindingFlags.NonPublic | BindingFlags.Public);
                var methodInfo = prop?.GetValue(function) as MethodInfo;
                return methodInfo?.GetCustomAttribute<ToolRiskAttribute>();
            }
            return null;
        }
    }
}
