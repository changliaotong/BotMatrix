using System.Reflection;

namespace BotWorker.Infrastructure.Utils
{
    public class PlaceholderContext
    {
        // 占位符名称 -> 解析函数（上下文传入）
        private readonly Dictionary<string, Func<Task<string>>> _asyncHandlers = [];
        private readonly Dictionary<string, Func<string>> _syncHandlers = [];
        // 占位符描述
        private readonly Dictionary<string, string> _descriptions = [];
        // 是否启用
        private readonly Dictionary<string, bool> _enabled = [];

        // 支持分组管理
        private readonly Dictionary<string, HashSet<string>> _groups = [];

        // 占位符匹配正则，支持默认值形式 {name|default}
        public string Pattern { get; set; } = @"\{(?<key>[^\}:|]+)(\|(?<default>[^\}]+))?\}";

        /// <summary>
        /// 注册异步占位符函数
        /// </summary>
        public void Register(string name, Func<Task<string>> asyncFunc, string description = "", bool enabled = true)
        {
            _asyncHandlers[name] = asyncFunc;
            _descriptions[name] = description;
            _enabled[name] = enabled;
        }

        /// <summary>
        /// 注册同步占位符函数，会自动包装成异步
        /// </summary>
        public void Register(string name, Func<string> syncFunc, string description = "", bool enabled = true)
        {
            _syncHandlers[name] = syncFunc;
            _descriptions[name] = description;
            _enabled[name] = enabled;
        }

        /// <summary>
        /// 获取占位符异步函数，可能为 null（禁用或未注册）
        /// </summary>
        public Func<Task<string>>? Get(string name)
        {
            if (_enabled.TryGetValue(name, out var enabled) && !enabled)
                return null;

            if (_asyncHandlers.TryGetValue(name, out var asyncFunc))
            {
                return asyncFunc;
            }
            else if (_syncHandlers.TryGetValue(name, out var syncFunc))
            {
                // 把同步函数包成异步函数返回
                return () => Task.FromResult(syncFunc());
            }

            return null;
        }


        // 启用/禁用占位符
        public void Enable(string name, bool enable = true)
        {
            if (_enabled.ContainsKey(name))
                _enabled[name] = enable;
        }

        // 启用/禁用某组
        public void EnableGroup(string group, bool enable = true)
        {
            if (_groups.ContainsKey(group))
            {
                foreach (var name in _groups[group])
                    _enabled[name] = enable;
            }
        }

        // 获取描述列表
        public Dictionary<string, string> ListDescriptions(string group = "")
        {
            if (string.IsNullOrEmpty(group))
                return new Dictionary<string, string>(_descriptions);

            var res = new Dictionary<string, string>();
            if (_groups.ContainsKey(group))
            {
                foreach (var name in _groups[group])
                {
                    if (_descriptions.TryGetValue(name, out var desc))
                        res[name] = desc;
                }
            }
            return res;
        }

        // 核心替换入口（异步，支持递归）
        public async Task ReplaceAsync(BotMessage bm, int maxDepth = 5)
        {
            if (string.IsNullOrEmpty(bm.Answer) || maxDepth <= 0)
                return;

            // 先替换条件语法 {if:cond?trueVal:falseVal}
            await ReplaceIfAsync(bm);

            // 再替换普通占位符 {key|default}
            await ReplacePlaceholdersAsync(bm);

            // 递归替换，避免无限循环，递归深度控制
            await ReplaceAsync(bm, maxDepth - 1);
        }

        public async Task ReplaceAsync(BotMessage bm)
        {
            string result = bm.Answer;
            int safetyCounter = 10;

            while (safetyCounter-- > 0 && Regex.IsMatch(result, Pattern))
            {
                // Step 1: 替换 if 条件表达式
                await ReplaceIfAsync(bm);

                // Step 2: 替换常规占位符
                var matches = Regex.Matches(bm.Answer, Pattern);
                var replacements = new Dictionary<string, string>();

                foreach (Match match in matches)
                {
                    string full = match.Groups[0].Value;
                    if (replacements.ContainsKey(full)) continue;

                    string name = match.Groups["key"].Value;
                    string? defaultValue = match.Groups["default"].Success ? match.Groups["default"].Value : null;
                    string replacement;

                    var func = Get(name);
                    if (func != null)
                    {
                        try
                        {
                            replacement = await func();
                        }
                        catch (Exception ex)
                        {
                            Console.WriteLine($"[占位符异常] {name}: {ex.Message}");
                            replacement = defaultValue ?? full;
                        }
                    }
                    else
                    {
                        replacement = defaultValue ?? full;
                    }

                    replacements[full] = replacement;
                }

                foreach (var pair in replacements)
                {
                    // InfoMessage($"[占位符] {pair.Key} -> {pair.Value}");
                    bm.Answer = bm.Answer.Replace(pair.Key, pair.Value);
                }
            }             
        }


        private static readonly string IfPattern = @"\{if:(?<cond>[^{}?]+)\?(?<trueVal>[^:{}]*)\:(?<falseVal>[^\}]+)\}";

        private static async Task<string> ReplaceIfAsync(BotMessage context)
        {
            return await Task.Run(() =>
            {
                return Regex.Replace(context.Answer, IfPattern, match =>
                {
                    string cond = match.Groups["cond"].Value.Trim();
                    string trueVal = match.Groups["trueVal"].Value;
                    string falseVal = match.Groups["falseVal"].Value;

                    bool result = EvaluateCondition(cond, context);
                    return result ? trueVal : falseVal;
                });
            });
        }

        private static bool EvaluateCondition(string condition, BotMessage context)
        {
            string[] ops = [">=", "<=", "!=", "==", "<>", "=", ">", "<"];

            foreach (var op in ops.OrderByDescending(o => o.Length))
            {
                int idx = condition.IndexOf(op);
                if (idx > 0)
                {
                    string left = condition[..idx].Trim();
                    string right = condition[(idx + op.Length)..].Trim();

                    var prop = typeof(BotMessage).GetProperty(left, BindingFlags.IgnoreCase | BindingFlags.Public | BindingFlags.Instance);
                    if (prop == null) return false;

                    var value = prop.GetValue(context.Answer);
                    string leftVal = value?.ToString() ?? "";

                    // Bool 特别处理
                    if (value is bool b)
                    {
                        return op switch
                        {
                            "=" or "==" => b.ToString().ToLower() == right.ToLower(),
                            "!=" or "<>" => b.ToString().ToLower() != right.ToLower(),
                            _ => false
                        };
                    }

                    // 数值判断
                    if (double.TryParse(leftVal, out double lv) && double.TryParse(right, out double rv))
                    {
                        return op switch
                        {
                            ">" => lv > rv,
                            "<" => lv < rv,
                            ">=" => lv >= rv,
                            "<=" => lv <= rv,
                            "=" or "==" => lv == rv,
                            "!=" or "<>" => lv != rv,
                            _ => false
                        };
                    }

                    // 字符串判断
                    return op switch
                    {
                        "=" or "==" => string.Equals(leftVal, right, StringComparison.OrdinalIgnoreCase),
                        "!=" or "<>" => !string.Equals(leftVal, right, StringComparison.OrdinalIgnoreCase),
                        _ => false
                    };
                }
            }

            // 简写支持：!IsOffical
            if (condition.StartsWith("!"))
            {
                var prop = typeof(BotMessage).GetProperty(condition[1..], BindingFlags.IgnoreCase | BindingFlags.Public | BindingFlags.Instance);
                if (prop?.PropertyType == typeof(bool))
                {
                    return !(bool)(prop.GetValue(context.Answer) ?? false);
                }
            }
            else
            {
                var prop = typeof(BotMessage).GetProperty(condition, BindingFlags.IgnoreCase | BindingFlags.Public | BindingFlags.Instance);
                if (prop?.PropertyType == typeof(bool))
                {
                    return (bool)(prop.GetValue(context.Answer) ?? false);
                }
            }

            return false;
        }

        // 替换普通占位符
        private async Task ReplacePlaceholdersAsync(BotMessage bm)
        {
            var regex = new Regex(Pattern, RegexOptions.IgnoreCase);

            var result = await Task.Run(async () =>
            {
                var matches = regex.Matches(bm.Answer);
                var output = bm.Answer;

                foreach (Match match in matches)
                {
                    string key = match.Groups["key"].Value;
                    string defaultVal = match.Groups["default"].Success ? match.Groups["default"].Value : "";

                    // 如果占位符未启用，跳过
                    if (!_enabled.ContainsKey(key) || !_enabled[key])
                        continue;

                    if (_asyncHandlers.TryGetValue(key, out var func))
                    {
                        string val = await func();
                        if (string.IsNullOrEmpty(val))
                            val = defaultVal;
                        if (!string.IsNullOrEmpty(val))
                            output = output.Replace(match.Value, val);
                    }

                    if (_syncHandlers.TryGetValue(key, out var syncFunc))
                    {
                        string val = syncFunc();
                        if (string.IsNullOrEmpty(val))
                            val = defaultVal;
                        if (!string.IsNullOrEmpty(val))
                            output = output.Replace(match.Value, val);
                    }
                }                
                
                return bm.Answer = output;
            });
        }
    }

}
