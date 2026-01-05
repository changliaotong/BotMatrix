using System;
using System.Collections.Generic;
using System.IO;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Threading.Tasks;
using System.Collections.Concurrent;
using System.Text.RegularExpressions;
using System.Linq;
using System.Text;
using System.Reflection;
using BotMatrix.SDK.Messaging;

namespace BotMatrix.SDK
{
    public class EventMessage
    {
        [JsonPropertyName("id")]
        public string Id { get; set; } = string.Empty;

        [JsonPropertyName("type")]
        public string Type { get; set; } = string.Empty;

        [JsonPropertyName("name")]
        public string Name { get; set; } = string.Empty;

        [JsonPropertyName("correlation_id")]
        public string? CorrelationId { get; set; }

        [JsonPropertyName("payload")]
        public Dictionary<string, object> Payload { get; set; } = new Dictionary<string, object>();
    }

    public class Action
    {
        [JsonPropertyName("type")]
        public string Type { get; set; } = string.Empty;

        [JsonPropertyName("target")]
        public string? Target { get; set; }

        [JsonPropertyName("target_id")]
        public string? TargetId { get; set; }

        [JsonPropertyName("text")]
        public string? Text { get; set; }

        [JsonPropertyName("correlation_id")]
        [JsonIgnore(Condition = JsonIgnoreCondition.WhenWritingNull)]
        public string? CorrelationId { get; set; }

        [JsonPropertyName("payload")]
        public Dictionary<string, object> Payload { get; set; } = new Dictionary<string, object>();
    }

    public class ResponseMessage
    {
        [JsonPropertyName("id")]
        public string Id { get; set; } = string.Empty;

        [JsonPropertyName("ok")]
        public bool Ok { get; set; }

        [JsonPropertyName("actions")]
        public List<Action> Actions { get; set; } = new List<Action>();

        [JsonPropertyName("error")]
        [JsonIgnore(Condition = JsonIgnoreCondition.WhenWritingNull)]
        public string? Error { get; set; }
    }

    public class Session
    {
        private readonly Context _ctx;
        private readonly BotMatrixPlugin _plugin;
        public Session(Context ctx, BotMatrixPlugin plugin)
        {
            _ctx = ctx;
            _plugin = plugin;
        }

        public void Set(string key, object value, int expireSeconds = 0)
        {
            _ctx.CallAction("storage.set", new Dictionary<string, object> {
                { "key", key },
                { "value", value },
                { "expire", expireSeconds }
            });
        }

        public async Task SetAsync(string key, object value, int expireSeconds = 0)
        {
            string correlationId = $"storage_set_{Guid.NewGuid()}";
            
            _plugin.RegisterWaitingSession(correlationId, new TaskCompletionSource<Context>()); 

            _ctx.CallAction("storage.set", new Dictionary<string, object> {
                { "key", key },
                { "value", value },
                { "expire", expireSeconds },
                { "correlation_id", correlationId }
            });

            await Task.CompletedTask;
        }

        public async Task<T> GetAsync<T>(string key, T defaultValue = default!)
        {
            string correlationId = $"storage_get_{Guid.NewGuid()}";
            var tcs = new TaskCompletionSource<Context>();
            _plugin.RegisterWaitingSession(correlationId, tcs);

            _ctx.CallAction("storage.get", new Dictionary<string, object> {
                { "key", key },
                { "correlation_id", correlationId }
            });

            try
            {
                var resultCtx = await tcs.Task.WaitAsync(TimeSpan.FromSeconds(5));
                if (resultCtx.Event.Payload.TryGetValue("value", out var val))
                {
                    if (val is JsonElement elem)
                    {
                        return JsonSerializer.Deserialize<T>(elem.GetRawText()) ?? defaultValue;
                    }
                    return (T)Convert.ChangeType(val, typeof(T));
                }
                return defaultValue;
            }
            catch
            {
                return defaultValue;
            }
            finally
            {
                _plugin.UnregisterWaitingSession(correlationId);
            }
        }

        public async Task DeleteAsync(string key)
        {
            _ctx.CallAction("storage.delete", new Dictionary<string, object> { { "key", key } });
            await Task.CompletedTask;
        }
    }

    public class Context : IMessageContext
    {
        public EventMessage Event { get; }
        public List<Action> Actions { get; } = new List<Action>();
        public string[] Args { get; internal set; } = Array.Empty<string>();
        public Dictionary<string, string> Params { get; internal set; } = new Dictionary<string, string>();
        
        public Session Session { get; }
        public string? Result { get; set; }
        public string Answer { get => Result ?? ""; set => Result = value; }

        public string UserId => Event.Payload.TryGetValue("from", out var val) ? val?.ToString() ?? "" : "";
        public string UserName => Event.Payload.TryGetValue("nickname", out var val) ? val?.ToString() ?? (Event.Payload.TryGetValue("sender", out var sender) && sender is JsonElement s && s.TryGetProperty("nickname", out var nick) ? nick.GetString() ?? "" : "") : "";
        public string GroupId => Event.Payload.TryGetValue("group_id", out var val) ? val?.ToString() ?? "" : "";
        public string Platform => Event.Payload.TryGetValue("platform", out var val) ? val?.ToString() ?? "" : "";
        public string MessageText => Event.Payload.TryGetValue("text", out var val) ? val?.ToString() ?? "" : "";
        public string GetBotId() => Event.Payload.TryGetValue("self_id", out var val) ? val?.ToString() ?? "" : "";
        public long BotId => long.TryParse(GetBotId(), out var id) ? id : 0;
        public string CommandName => Params.TryGetValue("command", out var val) ? val : "";
        public string ActionName => Params.TryGetValue("action", out var val) ? val : "";
        public string GroupName => Event.Payload.TryGetValue("group_name", out var val) ? val?.ToString() ?? "" : "";

        // 兼容性快捷属性
        public string NoticeType => Event.Payload.TryGetValue("notice_type", out var val) ? val?.ToString() ?? "" : "";
        public string SubType => Event.Payload.TryGetValue("sub_type", out var val) ? val?.ToString() ?? "" : "";
        public string TargetId => Event.Payload.TryGetValue("target_id", out var val) ? val?.ToString() ?? "" : "";
        public string SelfId => Event.Payload.TryGetValue("self_id", out var val) ? val?.ToString() ?? "" : "";

        private readonly BotMatrixPlugin _plugin;
        private readonly object _lock = new object();
        public Dictionary<string, Func<string>> Placeholders { get; } = new();

        public Context(EventMessage @event, BotMatrixPlugin plugin)
        {
            Event = @event;
            _plugin = plugin;
            Session = new Session(this, plugin);
        }

        public void Reply(string text)
        {
            if (string.IsNullOrEmpty(text)) return;
            SendText(text);
        }

        public void SendText(string text)
        {
            if (string.IsNullOrEmpty(text)) return;
            CallAction("send_message", new Dictionary<string, object> { { "text", text } });
        }

        public void RegisterPlaceholder(string key, Func<string> valueFactory)
        {
            Placeholders[key] = valueFactory;
        }

        public void SendImage(string url)
        {
            CallAction("send_image", new Dictionary<string, object> { { "url", url } });
        }

        public void SendAt(string userId, string text = "")
        {
            var payload = new Dictionary<string, object> { { "user_id", userId } };
            if (!string.IsNullOrEmpty(text)) payload["text"] = text;
            CallAction("send_at", payload);
        }

        public void SendFace(int id)
        {
            CallAction("send_face", new Dictionary<string, object> { { "id", id } });
        }

        public void SendRecord(string url)
        {
            CallAction("send_record", new Dictionary<string, object> { { "url", url } });
        }

        public void SendVideo(string url)
        {
            CallAction("send_video", new Dictionary<string, object> { { "url", url } });
        }

        public async Task SetGroupSpecialTitleAsync(string title, string? userId = null, string? groupId = null)
        {
            await CallActionAsync("set_group_special_title", new
            {
                group_id = groupId ?? GroupId,
                user_id = userId ?? UserId,
                special_title = title
            });
        }

        public async Task SetGroupCardAsync(string card, string? userId = null, string? groupId = null)
        {
            await CallActionAsync("set_group_card", new
            {
                group_id = groupId ?? GroupId,
                user_id = userId ?? UserId,
                card = card
            });
        }

        public async Task BanGroupMemberAsync(int durationSeconds, string? userId = null, string? groupId = null)
        {
            await CallActionAsync("set_group_ban", new
            {
                group_id = groupId ?? GroupId,
                user_id = userId ?? UserId,
                duration = durationSeconds
            });
        }

        public async Task KickGroupMemberAsync(bool rejectAddRequest = false, string? userId = null, string? groupId = null)
        {
            await CallActionAsync("set_group_kick", new
            {
                group_id = groupId ?? GroupId,
                user_id = userId ?? UserId,
                reject_add_request = rejectAddRequest
            });
        }

        public async Task SetGroupAdminAsync(bool enable, string? userId = null, string? groupId = null)
        {
            await CallActionAsync("set_group_admin", new
            {
                group_id = groupId ?? GroupId,
                user_id = userId ?? UserId,
                enable = enable
            });
        }

        public async Task SetGroupWholeBanAsync(bool enable, string? groupId = null)
        {
            await CallActionAsync("set_group_whole_ban", new
            {
                group_id = groupId ?? GroupId,
                enable = enable
            });
        }

        public async Task<Context> AskAsync(string prompt, int timeoutMs = 30000)
        {
            string correlationId = $"ask_{Event.Id}_{Guid.NewGuid()}";
            
            CallAction("send_message", new Dictionary<string, object> { 
                { "text", prompt },
                { "correlation_id", correlationId }
            });

            var tcs = new TaskCompletionSource<Context>();
            _plugin.RegisterWaitingSession(correlationId, tcs);

            try
            {
                return await tcs.Task.WaitAsync(TimeSpan.FromMilliseconds(timeoutMs));
            }
            catch (TimeoutException)
            {
                throw new TimeoutException($"用户在 {timeoutMs}ms 内未响应。");
            }
            finally
            {
                _plugin.UnregisterWaitingSession(correlationId);
            }
        }

        public async Task<JsonElement?> CallSkillAsync(string pluginId, string skillName, object? payload = null)
        {
            return await CallSkillAsync<JsonElement?>(pluginId, skillName, payload);
        }

        public async Task<T?> CallSkillAsync<T>(string pluginId, string skillName, object? payload = null)
        {
            string correlationId = $"skill_{skillName}_{Guid.NewGuid()}";
            var tcs = new TaskCompletionSource<Context>();
            _plugin.RegisterWaitingSession(correlationId, tcs);

            var skillPayload = new Dictionary<string, object> {
                { "plugin_id", pluginId },
                { "skill", skillName },
                { "correlation_id", correlationId }
            };

            if (payload != null)
            {
                var json = JsonSerializer.Serialize(payload);
                var dict = JsonSerializer.Deserialize<Dictionary<string, object>>(json);
                if (dict != null)
                {
                    foreach (var kv in dict) skillPayload[kv.Key] = kv.Value;
                }
            }

            CallAction("call_skill", skillPayload);

            try
            {
                var resultCtx = await tcs.Task.WaitAsync(TimeSpan.FromSeconds(10));
                if (resultCtx.Event.Payload.TryGetValue("result", out var result))
                {
                    if (result is JsonElement elem)
                    {
                        return JsonSerializer.Deserialize<T>(elem.GetRawText());
                    }
                    return (T?)Convert.ChangeType(result, typeof(T));
                }
                return default;
            }
            finally
            {
                _plugin.UnregisterWaitingSession(correlationId);
            }
        }

        public async Task<JsonElement?> CallActionAsync(string actionType, object? parameters = null)
        {
            return await CallActionAsync<JsonElement?>(actionType, parameters);
        }

        public async Task<T?> CallActionAsync<T>(string actionType, object? parameters = null)
        {
            string correlationId = $"{actionType}_{Guid.NewGuid()}";
            var tcs = new TaskCompletionSource<Context>();
            _plugin.RegisterWaitingSession(correlationId, tcs);

            var payload = new Dictionary<string, object>();
            if (parameters != null)
            {
                if (parameters is Dictionary<string, object> dict)
                {
                    payload = dict;
                }
                else
                {
                    var json = JsonSerializer.Serialize(parameters);
                    var d = JsonSerializer.Deserialize<Dictionary<string, object>>(json);
                    if (d != null) payload = d;
                }
            }
            payload["correlation_id"] = correlationId;

            CallAction(actionType, payload);

            try
            {
                var resultCtx = await tcs.Task.WaitAsync(TimeSpan.FromSeconds(10));
                if (resultCtx.Event.Payload.TryGetValue("data", out var data))
                {
                    if (data is JsonElement elem)
                    {
                        return JsonSerializer.Deserialize<T>(elem.GetRawText());
                    }
                    return (T?)Convert.ChangeType(data, typeof(T));
                }
                // Fallback: check top-level payload if "data" not present
                if (resultCtx.Event.Payload.Count > 0)
                {
                    var json = JsonSerializer.Serialize(resultCtx.Event.Payload);
                    return JsonSerializer.Deserialize<T>(json);
                }
                return default;
            }
            finally
            {
                _plugin.UnregisterWaitingSession(correlationId);
            }
        }

        public void ReplyImage(string url)
        {
            CallAction("send_image", new Dictionary<string, object> { { "url", url } });
        }

        public void DeleteMessage(string messageId)
        {
            CallAction("delete_message", new Dictionary<string, object> { { "message_id", messageId } });
        }

        public void KickUser(string groupId, string userId)
        {
            CallAction("kick_user", new Dictionary<string, object> { { "group_id", groupId }, { "user_id", userId } });
        }

        public void AddAction(string name, object? payload = null)
        {
            if (payload == null)
            {
                CallAction(name, null);
            }
            else
            {
                var json = JsonSerializer.Serialize(payload);
                var dict = JsonSerializer.Deserialize<Dictionary<string, object>>(json);
                CallAction(name, dict);
            }
        }

        public void CallAction(string actionType, Dictionary<string, object>? parameters = null)
        {
            // Permission check against plugin.json 'actions'
            if (!_plugin.HasPermission(actionType))
            {
                Console.Error.WriteLine($"[SDK] Permission denied: Action '{actionType}' is not declared in plugin.json");
                return;
            }

            lock (_lock)
            {
                var from = Event.Payload.TryGetValue("from", out var f) ? f?.ToString() : null;
                var groupId = Event.Payload.TryGetValue("group_id", out var g) ? g?.ToString() : null;
                var platform = Platform;
                var selfId = SelfId;

                var action = new Action
                {
                    Type = actionType,
                    Target = from,
                    TargetId = groupId,
                    Payload = parameters ?? new Dictionary<string, object>()
                };

                // Automatically include platform and self_id if not already present in parameters
                if (!string.IsNullOrEmpty(platform) && !action.Payload.ContainsKey("platform"))
                {
                    action.Payload["platform"] = platform;
                }
                if (!string.IsNullOrEmpty(selfId) && !action.Payload.ContainsKey("self_id"))
                {
                    action.Payload["self_id"] = selfId;
                }

                if (parameters != null)
                {
                    if (parameters.TryGetValue("text", out var textVal))
                    {
                        action.Text = textVal?.ToString();
                    }
                    if (parameters.TryGetValue("correlation_id", out var corrVal))
                    {
                        action.CorrelationId = corrVal?.ToString();
                    }
                }

                Actions.Add(action);
            }
        }
    }

    public delegate Task HandlerDelegate(Context ctx);
    public delegate HandlerDelegate MiddlewareDelegate(HandlerDelegate next);

    public class BotMatrixPlugin
    {
        private readonly ConcurrentDictionary<string, List<HandlerDelegate>> _handlers = new ConcurrentDictionary<string, List<HandlerDelegate>>();
        private readonly List<MiddlewareDelegate> _middlewares = new List<MiddlewareDelegate>();
        private readonly List<string> _registeredCommands = new List<string>();

        public IEnumerable<string> RegisteredCommands => _registeredCommands.AsReadOnly();
        private readonly BlockingCollection<ResponseMessage> _outputQueue = new BlockingCollection<ResponseMessage>();
        private readonly ConcurrentDictionary<string, TaskCompletionSource<Context>> _waitingSessions = new ConcurrentDictionary<string, TaskCompletionSource<Context>>();
        private JsonElement? _config;
        public JsonElement? Config => _config;

        public BotMatrixPlugin()
        {
            LoadConfig("plugin.json");
        }

        private void LoadConfig(string path)
        {
            if (File.Exists(path))
            {
                try
                {
                    var json = File.ReadAllText(path);
                    using var doc = JsonDocument.Parse(json);
                    _config = doc.RootElement.Clone();
                }
                catch { }
            }
        }

        public bool HasPermission(string action)
        {
            if (_config == null) return true; // Legacy mode

            // 1. Check 'permissions' (Actions this plugin is allowed to call)
            if (_config.Value.TryGetProperty("permissions", out var perms) && perms.ValueKind == JsonValueKind.Array)
            {
                foreach (var item in perms.EnumerateArray())
                {
                    if (item.GetString() == action) return true;
                }
            }

            // 2. Check 'actions' (Actions this plugin exports - usually can also call its own actions)
            if (_config.Value.TryGetProperty("actions", out var actions) && actions.ValueKind == JsonValueKind.Array)
            {
                foreach (var item in actions.EnumerateArray())
                {
                    if (item.ValueKind == JsonValueKind.String)
                    {
                        if (item.GetString() == action) return true;
                    }
                    else if (item.ValueKind == JsonValueKind.Object)
                    {
                        if (item.TryGetProperty("name", out var nameProp) && nameProp.GetString() == action) return true;
                    }
                }
            }

            // Special case for built-in essential actions if not declared
            if (action == "send_message" || action == "storage.get" || action == "storage.set") return true;

            return false;
        }

        public void RegisterWaitingSession(string key, TaskCompletionSource<Context> tcs)
        {
            _waitingSessions[key] = tcs;
        }

        public void UnregisterWaitingSession(string key)
        {
            _waitingSessions.TryRemove(key, out _);
        }

        public void Use(MiddlewareDelegate middleware)
        {
            _middlewares.Add(middleware);
        }

        public void On(string eventName, HandlerDelegate handler)
        {
            _handlers.AddOrUpdate(eventName, 
                new List<HandlerDelegate> { handler }, 
                (key, existing) => {
                    lock(existing) {
                        existing.Add(handler);
                    }
                    return existing;
                });
        }

        public void OnMessage(HandlerDelegate handler)
        {
            On("on_message", handler);
        }

        public void OnNotice(HandlerDelegate handler)
        {
            On("on_notice", handler);
        }

        public void OnIntent(string intentName, HandlerDelegate handler)
        {
            On($"intent_{intentName}", handler);
        }

        public void Command(string[] aliases, HandlerDelegate handler)
        {
            _registeredCommands.AddRange(aliases);
            OnMessage(async ctx =>
            {
                var text = ctx.Event.Payload.ContainsKey("text") ? ctx.Event.Payload["text"]?.ToString() : "";
                if (string.IsNullOrWhiteSpace(text)) return;

                // 1. 处理前缀 / (可选) 和 空格
                // 正则说明：可选的斜杠 ^/?，后跟任意个空白字符 \s*，然后匹配别名
                // 支持别名后紧跟数字的情况（如 c100），通过正向先行断言 (?=\d) 实现
                foreach (var alias in aliases)
                {
                    var pattern = $@"^/?\s*{Regex.Escape(alias)}(\s+|(?=\d)|$)";
                    var match = Regex.Match(text, pattern, RegexOptions.IgnoreCase);
                    
                    if (match.Success)
                    {
                        ctx.Params["command"] = alias;
                        var remaining = text.Substring(match.Index + match.Length).Trim();
                        ctx.Args = remaining.Split(new[] { ' ', '\u3000' }, StringSplitOptions.RemoveEmptyEntries);
                        await handler(ctx);
                        return;
                    }
                }
            });
        }

        public void Command(string cmd, HandlerDelegate handler)
        {
            Command(new[] { cmd }, handler);
        }

        public void RegexCommand(string pattern, HandlerDelegate handler)
        {
            var regex = new Regex(pattern);
            OnMessage(async ctx =>
            {
                var text = ctx.Event.Payload.ContainsKey("text") ? ctx.Event.Payload["text"]?.ToString() : "";
                if (text == null) return;

                var match = regex.Match(text);
                if (match.Success)
                {
                    ctx.Args = match.Groups.Cast<Group>().Select(g => g.Value).ToArray();
                    ctx.Params = regex.GetGroupNames()
                        .Where(name => !int.TryParse(name, out _))
                        .ToDictionary(name => name, name => match.Groups[name].Value);
                    await handler(ctx);
                }
            });
        }

        public void ExportSkill(string name, HandlerDelegate handler)
        {
            On("skill_" + name, handler);
        }

        public void ExportSkill<TIn, TOut>(string name, Func<Context, TIn, Task<TOut>> handler)
        {
            ExportSkill(name, async ctx => {
                TIn input = default;
                if (ctx.Event.Payload.TryGetValue("params", out var p))
                {
                    if (p is JsonElement elem) input = JsonSerializer.Deserialize<TIn>(elem.GetRawText());
                    else input = (TIn)Convert.ChangeType(p, typeof(TIn));
                }
                
                var result = await handler(ctx, input);
                ctx.Actions.Add(new Action {
                    Type = "skill_result",
                    Payload = new Dictionary<string, object> { { "result", result } }
                });
            });
        }

        public void OnAction(string name, HandlerDelegate handler)
        {
            ExportSkill(name, handler);
        }

        public async Task EmitAction(string name, Dictionary<string, object> payload)
        {
            if (_handlers.TryGetValue("skill_" + name, out var handlers))
            {
                var msg = new EventMessage { Name = "skill_" + name, Payload = payload, Type = "event" };
                var ctx = new Context(msg, this);
                foreach (var handler in handlers)
                {
                    await handler(ctx);
                }
            }
        }

        public async Task EmitIntent(string name, Dictionary<string, object> payload)
        {
            if (_handlers.TryGetValue("intent_" + name, out var handlers))
            {
                var msg = new EventMessage { Name = "intent_" + name, Payload = payload, Type = "event" };
                var ctx = new Context(msg, this);
                foreach (var handler in handlers)
                {
                    await handler(ctx);
                }
            }
        }

        public async Task RunAsync()
        {
            Console.OutputEncoding = new System.Text.UTF8Encoding(false); // Force UTF8 without BOM
            Console.InputEncoding = new System.Text.UTF8Encoding(false);

            // Start output worker
            _ = Task.Run(() =>
            {
                foreach (var resp in _outputQueue.GetConsumingEnumerable())
                {
                    Console.WriteLine(JsonSerializer.Serialize(resp));
                }
            });

            using (var reader = new StreamReader(Console.OpenStandardInput(), new System.Text.UTF8Encoding(false)))
            {
                while (true)
                {
                    var line = await reader.ReadLineAsync();
                    if (line == null) break;
                    if (string.IsNullOrWhiteSpace(line)) continue; // Skip empty lines

                    try
                    {
                        var msg = JsonSerializer.Deserialize<EventMessage>(line);
                        if (msg?.Type == "event")
                        {
                            // Spawn concurrent task for each event
                            _ = HandleEventAsync(msg);
                        }
                    }
                    catch (Exception ex)
                    {
                        Console.Error.WriteLine($"[SDK] Error deserializing message: {ex.Message}");
                    }
                }
            }
            _outputQueue.CompleteAdding();
        }

        private async Task HandleEventAsync(EventMessage msg)
        {
            // 1. Check by CorrelationId first (The most reliable way in distributed systems)
            if (!string.IsNullOrEmpty(msg.CorrelationId) && _waitingSessions.TryGetValue(msg.CorrelationId, out var tcsById))
            {
                tcsById.SetResult(new Context(msg, this));
                _outputQueue.Add(new ResponseMessage { Id = msg.Id, Ok = true, Actions = new List<Action>() });
                return;
            }

            // 2. Fallback to session key (for local backward compatibility)
            if (msg.Name == "on_message")
            {
                string from = msg.Payload.ContainsKey("from") ? msg.Payload["from"]?.ToString() : "";
                string groupId = msg.Payload.ContainsKey("group_id") ? msg.Payload["group_id"]?.ToString() : "";
                string sessionKey = $"{groupId}:{from}";

                if (_waitingSessions.TryGetValue(sessionKey, out var tcs))
                {
                    tcs.SetResult(new Context(msg, this));
                    _outputQueue.Add(new ResponseMessage { Id = msg.Id, Ok = true, Actions = new List<Action>() });
                    return;
                }
            }

            if (!_handlers.TryGetValue(msg.Name, out var handlers))
            {
                _outputQueue.Add(new ResponseMessage { Id = msg.Id, Ok = true, Actions = new List<Action>() });
                return;
            }

            // Wrap with middleware
            HandlerDelegate finalHandler = async ctx => {
                foreach (var handler in handlers)
                {
                    await handler(ctx);
                }
            };

            for (int i = _middlewares.Count - 1; i >= 0; i--)
            {
                finalHandler = _middlewares[i](finalHandler);
            }

            var ctx = new Context(msg, this);
            try
            {
                await finalHandler(ctx);
                
                // If there's a result string, send it as a reply
                if (!string.IsNullOrEmpty(ctx.Result))
                {
                    ctx.Reply(ctx.Result);
                }

                // Centralized FriendlyMessage processing
                var fm = new FriendlyMessage(ctx);
                foreach (var action in ctx.Actions)
                {
                    if (!string.IsNullOrEmpty(action.Text))
                    {
                        action.Text = await fm.GetFriendlyResAsync(action.Text);
                        
                        // Sync back to payload if it exists there
                        if (action.Payload.ContainsKey("text"))
                        {
                            action.Payload["text"] = action.Text;
                        }
                    }
                }

                _outputQueue.Add(new ResponseMessage { Id = msg.Id, Ok = true, Actions = ctx.Actions });
            }
            catch (Exception ex)
            {
                Console.Error.WriteLine($"[SDK] Handler error for {msg.Name}: {ex.Message}");
                _outputQueue.Add(new ResponseMessage { Id = msg.Id, Ok = false, Error = ex.Message });
            }
        }

        public void Run()
        {
            RunAsync().GetAwaiter().GetResult();
        }
    }

    /// <summary>
    /// 提供更方便的插件本地调试工具
    /// </summary>
    public class PluginDebugger
    {
        private readonly BotMatrixPlugin _plugin;
        private string _selfId = "51437810";
        private string _currentUserId = "1653346663";
        private string _currentUserName = "光辉岁月";
        private string _currentGroupId = "86433316";
        private string _currentGroupName = "测试群组";
        private string _platform = "qq";
        private readonly List<string> _history = new List<string>();

        public PluginDebugger(BotMatrixPlugin plugin)
        {
            _plugin = plugin;
        }

        private int _historyIndex = -1;

        public async Task StartAsync(string[]? args = null)
        {
            Console.Clear();
            PrintHeader();

            // 检查命令行参数是否包含自动化测试请求
            if (args != null && args.Length >= 2 && args[0].Equals("--test", StringComparison.OrdinalIgnoreCase))
            {
                var testFileName = args[1];
                string filePath;
                
                if (Path.IsPathRooted(testFileName))
                {
                    filePath = testFileName;
                }
                else
                {
                    // 1. 尝试直接作为相对 CWD 的路径
                    filePath = Path.GetFullPath(testFileName);
                    if (!File.Exists(filePath))
                    {
                        // 2. 尝试作为相对 BaseDirectory 的路径
                        filePath = Path.GetFullPath(Path.Combine(AppDomain.CurrentDomain.BaseDirectory, testFileName));
                        
                        if (!File.Exists(filePath) && !testFileName.StartsWith("tests", StringComparison.OrdinalIgnoreCase))
                        {
                            // 3. 尝试在 BaseDirectory/tests 目录下查找
                            var testDir = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "tests");
                            filePath = Path.Combine(testDir, testFileName.EndsWith(".txt") ? testFileName : testFileName + ".txt");
                        }
                    }
                }
                
                if (File.Exists(filePath))
                {
                    await RunBatchTestFileAsync(filePath);
                    Console.WriteLine("\n自动化测试完成，按任意键退出或输入 /exit ...");
                }
                else
                {
                    Console.WriteLine($"\n❌ 找不到测试文件: {filePath}");
                }
            }

            while (true)
            {
                PrintPrompt();
                string input = await ReadLineWithHistoryAsync();

                if (string.IsNullOrWhiteSpace(input)) continue;

                if (input.StartsWith("/"))
                {
                    if (HandleSystemCommand(input)) break;
                    continue;
                }

                _history.Add(input);
                _historyIndex = -1;
                await SimulateMessageAsync(input);
            }
        }

        private async Task<string> ReadLineWithHistoryAsync()
        {
            string currentInput = "";
            int cursorPosition = 0;

            while (true)
            {
                var key = Console.ReadKey(true);

                if (key.Key == ConsoleKey.Enter)
                {
                    Console.WriteLine();
                    return currentInput;
                }
                else if (key.Key == ConsoleKey.UpArrow)
                {
                    if (_history.Count > 0)
                    {
                        if (_historyIndex == -1) _historyIndex = _history.Count - 1;
                        else if (_historyIndex > 0) _historyIndex--;

                        ClearCurrentLine(currentInput.Length);
                        currentInput = _history[_historyIndex];
                        Console.Write(currentInput);
                        cursorPosition = currentInput.Length;
                    }
                }
                else if (key.Key == ConsoleKey.DownArrow)
                {
                    if (_historyIndex != -1)
                    {
                        if (_historyIndex < _history.Count - 1)
                        {
                            _historyIndex++;
                            currentInput = _history[_historyIndex];
                        }
                        else
                        {
                            _historyIndex = -1;
                            currentInput = "";
                        }

                        ClearCurrentLine(currentInput.Length);
                        Console.Write(currentInput);
                        cursorPosition = currentInput.Length;
                    }
                }
                else if (key.Key == ConsoleKey.Tab)
                {
                    var suggestions = GetSuggestions(currentInput);
                    if (suggestions.Any())
                    {
                        ClearCurrentLine(currentInput.Length);
                        currentInput = suggestions.First();
                        Console.Write(currentInput);
                        cursorPosition = currentInput.Length;
                    }
                }
                else if (key.Key == ConsoleKey.Backspace)
                {
                    if (cursorPosition > 0)
                    {
                        currentInput = currentInput.Remove(cursorPosition - 1, 1);
                        cursorPosition--;
                        Console.Write("\b \b");
                        // If we are not at the end, we'd need more complex logic, but for simplicity:
                        if (cursorPosition < currentInput.Length)
                        {
                            var rest = currentInput.Substring(cursorPosition);
                            Console.Write(rest + " ");
                            for (int i = 0; i <= rest.Length; i++) Console.Write("\b");
                        }
                    }
                }
                else if (key.KeyChar >= 32)
                {
                    currentInput = currentInput.Insert(cursorPosition, key.KeyChar.ToString());
                    Console.Write(key.KeyChar);
                    cursorPosition++;
                    if (cursorPosition < currentInput.Length)
                    {
                        var rest = currentInput.Substring(cursorPosition);
                        Console.Write(rest);
                        for (int i = 0; i < rest.Length; i++) Console.Write("\b");
                    }
                }
            }
        }

        private List<string> GetSuggestions(string input)
        {
            if (string.IsNullOrWhiteSpace(input)) return new List<string>();

            var allPossible = _plugin.RegisteredCommands.ToList();
            allPossible.AddRange(new[] { "/help", "/user", "/group", "/test", "/clear", "/exit" });

            return allPossible
                .Where(c => c.StartsWith(input, StringComparison.OrdinalIgnoreCase))
                .OrderBy(c => c.Length)
                .ToList();
        }

        private void ClearCurrentLine(int currentLength)
        {
            // Move cursor to start of input (after prompt)
            // This is a bit tricky with just Console.Write("\b")
            // A better way is to know where the prompt ended.
            // But since we just want to clear the line:
            for (int i = 0; i < currentLength; i++) Console.Write("\b \b");
        }

        private void PrintHeader()
        {
            Console.ForegroundColor = ConsoleColor.Cyan;
            Console.WriteLine("==================================================");
            Console.WriteLine("        BotMatrix 插件本地调试控制台        ");
            Console.WriteLine("==================================================");
            Console.ForegroundColor = ConsoleColor.Gray;
            Console.WriteLine("输入消息进行测试，输入 /help 查看系统指令");
            Console.WriteLine($"当前环境: 用户:{_currentUserName}({_currentUserId}) | 群组:{_currentGroupName}({_currentGroupId})");
            Console.WriteLine("--------------------------------------------------");
            Console.ResetColor();
        }

        private void PrintPrompt()
        {
            Console.ForegroundColor = ConsoleColor.Green;
            Console.Write($"{_currentUserName}> ");
            Console.ResetColor();
        }

        private async Task ReadBatchTestsAsync()
        {
            var testDir = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "tests");
            if (!Directory.Exists(testDir))
            {
                Console.WriteLine("⚠️ 未找到测试包目录 (tests/)");
                return;
            }

            var files = Directory.GetFiles(testDir, "*.txt");
            if (files.Length == 0)
            {
                Console.WriteLine("⚠️ tests/ 目录下没有 .txt 测试文件");
                return;
            }

            Console.WriteLine("\n可用测试包:");
            for (int i = 0; i < files.Length; i++)
            {
                Console.WriteLine($"  [{i + 1}] {Path.GetFileName(files[i])}");
            }
            Console.WriteLine("  [0] 取消");

            Console.Write("\n选择测试包编号: ");
            if (int.TryParse(Console.ReadLine(), out int choice) && choice > 0 && choice <= files.Length)
            {
                await RunBatchTestFileAsync(files[choice - 1]);
            }
        }

        private async Task RunBatchTestFileAsync(string filePath)
        {
            Console.WriteLine($"\n🚀 开始执行测试包: {Path.GetFileName(filePath)}");
            var lines = await File.ReadAllLinesAsync(filePath);
            int success = 0;
            int total = 0;

            foreach (var line in lines)
            {
                var trimmed = line.Trim();
                if (string.IsNullOrWhiteSpace(trimmed) || trimmed.StartsWith("#")) continue;

                total++;
                Console.ForegroundColor = ConsoleColor.Yellow;
                Console.WriteLine($"\n[测试 {total}] 输入: {trimmed}");
                Console.ResetColor();

                if (trimmed.StartsWith("/"))
                {
                    HandleSystemCommand(trimmed);
                }
                else
                {
                    await SimulateMessageAsync(trimmed);
                }
                success++;
                
                // 稍微停顿一下，方便观察
                await Task.Delay(500);
            }

            Console.ForegroundColor = ConsoleColor.Green;
            Console.WriteLine($"\n✅ 测试完成! 成功执行 {success}/{total} 条指令。");
            Console.ResetColor();
        }

        private bool HandleSystemCommand(string input)
        {
            var parts = input.Split(' ', StringSplitOptions.RemoveEmptyEntries);
            var cmd = parts[0].ToLower();

            switch (cmd)
            {
                case "/help":
                    Console.WriteLine("系统指令列表:");
                    Console.WriteLine("  /user {id} {name}  - 切换模拟用户");
                    Console.WriteLine("  /group {id} {name} - 切换模拟群组");
                    Console.WriteLine("  /test              - 列出并运行批量测试包 (tests/*.txt)");
                    Console.WriteLine("  /clear             - 清屏");
                    Console.WriteLine("  /exit              - 退出调试");
                    break;
                case "/test":
                    ReadBatchTestsAsync().GetAwaiter().GetResult();
                    break;
                case "/user":
                    if (parts.Length >= 2) _currentUserId = parts[1];
                    if (parts.Length >= 3) _currentUserName = parts[2];
                    Console.WriteLine($"已切换用户: {_currentUserName}({_currentUserId})");
                    break;
                case "/group":
                    if (parts.Length >= 2) _currentGroupId = parts[1];
                    if (parts.Length >= 3) _currentGroupName = parts[2];
                    Console.WriteLine($"已切换群组: {_currentGroupName}({_currentGroupId})");
                    break;
                case "/bot":
                    if (parts.Length >= 2) _selfId = parts[1];
                    Console.WriteLine($"已切换机器人 ID: {_selfId}");
                    break;
                case "/clear":
                    Console.Clear();
                    PrintHeader();
                    break;
                case "/exit":
                    return true;
                default:
                    Console.WriteLine("未知指令，输入 /help 查看帮助");
                    break;
            }
            return false;
        }

        private async Task SimulateMessageAsync(string text)
        {
            var msg = new EventMessage
            {
                Id = Guid.NewGuid().ToString(),
                Name = "on_message",
                Type = "event",
                Payload = new Dictionary<string, object>
                {
                    { "from", _currentUserId },
                    { "nickname", _currentUserName },
                    { "group_id", _currentGroupId },
                    { "group_name", _currentGroupName },
                    { "text", text },
                    { "platform", _platform },
                    { "self_id", _selfId }
                }
            };

            // 获取插件内部的 _outputQueue 字段
            var outputQueueField = _plugin.GetType().GetField("_outputQueue", System.Reflection.BindingFlags.NonPublic | System.Reflection.BindingFlags.Instance);
            if (outputQueueField == null)
            {
                Console.WriteLine("[错误] 插件不兼容：找不到 _outputQueue 字段");
                return;
            }
            var queue = outputQueueField.GetValue(_plugin) as System.Collections.Concurrent.BlockingCollection<ResponseMessage>;

            // 拦截控制台输出，防止干扰
            var originalWriter = Console.Out;
            using (var sw = new StringWriter())
            {
                Console.SetOut(sw);
                
                try 
                {
                    var task = _plugin.GetType()
                        .GetMethod("HandleEventAsync", System.Reflection.BindingFlags.NonPublic | System.Reflection.BindingFlags.Instance)!
                        .Invoke(_plugin, new object[] { msg }) as Task;
                    
                    if (task != null) await task;
                }
                catch (Exception ex)
                {
                    Console.SetOut(originalWriter);
                    Console.ForegroundColor = ConsoleColor.Red;
                    Console.WriteLine($"[内部错误] {ex.InnerException?.Message ?? ex.Message}");
                    Console.ResetColor();
                }

                Console.SetOut(originalWriter);

                // 从队列中提取并显示结果
                if (queue != null)
                {
                    bool foundResponse = false;
                    // 我们需要给插件一点时间来生成响应消息，虽然 HandleEventAsync 应该已经完成了
                    // 但为了保险，我们可以检查队列
                    while (queue.TryTake(out var resp))
                    {
                        if (resp.Id == msg.Id)
                        {
                            PrintResponse(resp);
                            foundResponse = true;
                        }
                    }
                    
                    if (!foundResponse)
                    {
                        Console.ForegroundColor = ConsoleColor.DarkGray;
                        Console.WriteLine("[调试] 插件未对此消息生成响应内容。");
                        Console.ResetColor();
                    }
                }
            }
        }

        private void PrintResponse(ResponseMessage resp)
        {
            bool hasActions = false;
            if (resp.Actions != null && resp.Actions.Count > 0)
            {
                hasActions = true;
                foreach (var action in resp.Actions)
                {
                    if (action.Type == "send_message")
                    {
                        Console.ForegroundColor = ConsoleColor.Cyan;
                        Console.WriteLine($"[机器人] {action.Text}");
                        Console.ResetColor();
                    }
                    else
                    {
                        Console.ForegroundColor = ConsoleColor.DarkGray;
                        Console.WriteLine($"[Action] {action.Type}: {JsonSerializer.Serialize(action.Payload)}");
                        Console.ResetColor();
                    }
                }
            }

            if (!resp.Ok && !string.IsNullOrEmpty(resp.Error))
            {
                Console.ForegroundColor = ConsoleColor.Red;
                Console.WriteLine($"[错误] {resp.Error}");
                Console.ResetColor();
            }
            else if (!hasActions && resp.Ok)
            {
                // 如果没有动作但执行成功，给个反馈
                Console.ForegroundColor = ConsoleColor.DarkGray;
                Console.WriteLine("[系统] 指令已处理，无返回消息。");
                Console.ResetColor();
            }
        }
    }
}
