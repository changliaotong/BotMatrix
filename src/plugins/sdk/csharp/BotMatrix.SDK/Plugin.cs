using System;
using System.Collections.Generic;
using System.IO;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Threading.Tasks;
using System.Collections.Concurrent;
using System.Text.RegularExpressions;
using System.Linq;

namespace BotMatrix.SDK
{
    public class EventMessage
    {
        [JsonPropertyName("id")]
        public string Id { get; set; }

        [JsonPropertyName("type")]
        public string Type { get; set; }

        [JsonPropertyName("name")]
        public string Name { get; set; }

        [JsonPropertyName("correlation_id")]
        public string CorrelationId { get; set; }

        [JsonPropertyName("payload")]
        public Dictionary<string, object> Payload { get; set; }
    }

    public class Action
    {
        [JsonPropertyName("type")]
        public string Type { get; set; }

        [JsonPropertyName("target")]
        public string Target { get; set; }

        [JsonPropertyName("target_id")]
        public string TargetId { get; set; }

        [JsonPropertyName("text")]
        public string Text { get; set; }

        [JsonPropertyName("correlation_id")]
        [JsonIgnore(Condition = JsonIgnoreCondition.WhenWritingNull)]
        public string CorrelationId { get; set; }
    }

    public class ResponseMessage
    {
        [JsonPropertyName("id")]
        public string Id { get; set; }

        [JsonPropertyName("ok")]
        public bool Ok { get; set; }

        [JsonPropertyName("actions")]
        public List<Action> Actions { get; set; }

        [JsonPropertyName("error")]
        [JsonIgnore(Condition = JsonIgnoreCondition.WhenWritingNull)]
        public string Error { get; set; }
    }

    public class Context
    {
        public EventMessage Event { get; }
        public List<Action> Actions { get; } = new List<Action>();
        public string[] Args { get; internal set; } = Array.Empty<string>();
        public Dictionary<string, string> Params { get; internal set; } = new Dictionary<string, string>();
        
        private readonly BotMatrixPlugin _plugin;
        private readonly object _lock = new object();

        public Context(EventMessage @event, BotMatrixPlugin plugin)
        {
            Event = @event;
            _plugin = plugin;
        }

        public void Reply(string text)
        {
            CallAction("send_message", new Dictionary<string, object> { { "text", text } });
        }

        public async Task<Context> AskAsync(string prompt, int timeoutMs = 30000)
        {
            string correlationId = $"ask_{Event.Id}_{DateTimeOffset.UtcNow.ToUnixTimeMilliseconds()}";
            
            string from = Event.Payload.ContainsKey("from") ? Event.Payload["from"]?.ToString() : "";
            string groupId = Event.Payload.ContainsKey("group_id") ? Event.Payload["group_id"]?.ToString() : "";

            lock (_lock)
            {
                Actions.Add(new Action 
                { 
                    Type = "send_message", 
                    Target = from, 
                    TargetId = groupId, 
                    Text = prompt, 
                    CorrelationId = correlationId 
                });
            }

            var tcs = new TaskCompletionSource<Context>();
            _plugin.RegisterWaitingSession(correlationId, tcs);

            try
            {
                var delayTask = Task.Delay(timeoutMs);
                var completedTask = await Task.WhenAny(tcs.Task, delayTask);

                if (completedTask == tcs.Task)
                {
                    return await tcs.Task;
                }
                else
                {
                    throw new TimeoutException("The user did not respond within the timeout period.");
                }
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

        public void CallAction(string actionType, Dictionary<string, object> parameters = null)
        {
            // Permission check against plugin.json 'actions'
            if (!_plugin.HasPermission(actionType))
            {
                Console.Error.WriteLine($"[SDK] Permission denied: Action '{actionType}' is not declared in plugin.json");
                return;
            }

            lock (_lock)
            {
                string from = Event.Payload.ContainsKey("from") ? Event.Payload["from"]?.ToString() : "";
                string groupId = Event.Payload.ContainsKey("group_id") ? Event.Payload["group_id"]?.ToString() : "";

                var action = new Action
                {
                    Type = actionType,
                    Target = from,
                    TargetId = groupId
                };

                if (parameters != null && parameters.ContainsKey("text"))
                {
                    action.Text = parameters["text"]?.ToString();
                }

                Actions.Add(action);
            }
        }
    }

    public delegate Task HandlerDelegate(Context ctx);
    public delegate HandlerDelegate MiddlewareDelegate(HandlerDelegate next);

    public class BotMatrixPlugin
    {
        private readonly ConcurrentDictionary<string, HandlerDelegate> _handlers = new ConcurrentDictionary<string, HandlerDelegate>();
        private readonly List<MiddlewareDelegate> _middlewares = new List<MiddlewareDelegate>();
        private readonly BlockingCollection<ResponseMessage> _outputQueue = new BlockingCollection<ResponseMessage>();
        private readonly ConcurrentDictionary<string, TaskCompletionSource<Context>> _waitingSessions = new ConcurrentDictionary<string, TaskCompletionSource<Context>>();
        private JsonElement? _config;

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

            if (_config.Value.TryGetProperty("actions", out var actions) && actions.ValueKind == JsonValueKind.Array)
            {
                foreach (var item in actions.EnumerateArray())
                {
                    if (item.GetString() == action) return true;
                }
                return false;
            }
            return true;
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
            _handlers[eventName] = handler;
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

        public void Command(string cmd, HandlerDelegate handler)
        {
            OnMessage(async ctx =>
            {
                var text = ctx.Event.Payload.ContainsKey("text") ? ctx.Event.Payload["text"]?.ToString() : "";
                if (text != null && (text.StartsWith(cmd + " ") || text == cmd))
                {
                    ctx.Args = text.Substring(cmd.Length).Trim().Split(new[] { ' ' }, StringSplitOptions.RemoveEmptyEntries);
                    await handler(ctx);
                }
            });
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

        public async Task RunAsync()
        {
            // Start output worker
            _ = Task.Run(() =>
            {
                foreach (var resp in _outputQueue.GetConsumingEnumerable())
                {
                    Console.WriteLine(JsonSerializer.Serialize(resp));
                }
            });

            using (var reader = new StreamReader(Console.OpenStandardInput()))
            {
                while (true)
                {
                    var line = await reader.ReadLineAsync();
                    if (line == null) break;

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

            if (!_handlers.TryGetValue(msg.Name, out var handler))
            {
                _outputQueue.Add(new ResponseMessage { Id = msg.Id, Ok = true, Actions = new List<Action>() });
                return;
            }

            // Wrap with middleware
            var finalHandler = handler;
            for (int i = _middlewares.Count - 1; i >= 0; i--)
            {
                finalHandler = _middlewares[i](finalHandler);
            }

            var ctx = new Context(msg, this);
            try
            {
                await finalHandler(ctx);
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
}
