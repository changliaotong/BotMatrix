using System.Text.Json;
using StackExchange.Redis;
using BotWorker.Infrastructure.Communication.OneBot;


namespace BotWorker.Modules.Plugins
{
    public class PluginManager : IRobot
    {
        private readonly List<Skill> _skills = new();
        private readonly Dictionary<string, List<Func<IPluginContext, Task>>> _eventHandlers = new();
        private readonly List<IPlugin> _plugins = new();
        private readonly IAIService _aiService;
        private readonly IAgentExecutor _agentExecutor;
        private readonly II18nService _i18nService;
        private readonly IRagService _ragService;
        private readonly ILogger<PluginManager> _logger;
        private readonly IServiceProvider _serviceProvider;
        private readonly SessionManager _sessionManager;
        private readonly IEventNexus _eventNexus;
        private readonly IOneBotApiClient _oneBot;
        private IPlugin? _currentLoadingPlugin;

        private FileSystemWatcher? _watcher;
        private string? _lastScanDir;
        private readonly object _lock = new();
        private System.Timers.Timer? _reloadTimer;

        public IReadOnlyList<Skill> Skills => _skills;
        public IReadOnlyList<IPlugin> Plugins => _plugins;

        public IAIService AI => _aiService;
        public IAgentExecutor Agent => _agentExecutor;
        public II18nService I18n => _i18nService;
        public IRagService Rag => _ragService;
        public ILogger Logger => _logger;
        public SessionManager Sessions => _sessionManager;
        public IEventNexus Events => _eventNexus;

        public PluginManager(
            IAIService aiService, 
            IAgentExecutor agentExecutor,
            II18nService i18nService, 
            IRagService ragService,
            ILogger<PluginManager> logger,
            IServiceProvider serviceProvider,
            IConnectionMultiplexer redis,
            IEventNexus eventNexus,
            IOneBotApiClient oneBot)
        {
            _aiService = aiService;
            _agentExecutor = agentExecutor;
            _i18nService = i18nService;
            _ragService = ragService;
            _logger = logger;
            _serviceProvider = serviceProvider;
            _eventNexus = eventNexus;
            _oneBot = oneBot;
            _sessionManager = new SessionManager(redis);

            _reloadTimer = new System.Timers.Timer(3000); // 3秒防抖，与Go一致
            _reloadTimer.AutoReset = false;
            _reloadTimer.Elapsed += async (s, e) => await DoReloadAsync();
        }

        public async Task SendMessageAsync(string platform, string botId, string? groupId, string userId, string message)
        {
            _logger.LogInformation("[SendMessage] Platform: {Platform}, Bot: {BotId}, Group: {GroupId}, User: {UserId}, Message: {Message}", 
                platform, botId, groupId, userId, message);

            try
            {
                if (!string.IsNullOrEmpty(groupId))
                {
                    await _oneBot.SendActionAsync(platform, botId, "send_group_msg", new
                    {
                        group_id = groupId,
                        message = message
                    });
                }
                else
                {
                    await _oneBot.SendActionAsync(platform, botId, "send_private_msg", new
                    {
                        user_id = userId,
                        message = message
                    });
                }
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Failed to send message via OneBot API");
            }
        }

        public async Task ScanPluginsAsync(string baseDir)
        {
            _lastScanDir = baseDir;
            if (_watcher == null)
            {
                SetupWatcher(baseDir);
            }
            await FullScanAsync(baseDir);
        }

        private void SetupWatcher(string path)
        {
            if (!Directory.Exists(path)) return;
            
            _watcher = new FileSystemWatcher(path);
            _watcher.IncludeSubdirectories = true;
            _watcher.NotifyFilter = NotifyFilters.LastWrite | NotifyFilters.FileName | NotifyFilters.DirectoryName;
            _watcher.Filter = "*.*";

            _watcher.Changed += OnPluginDirChanged;
            _watcher.Created += OnPluginDirChanged;
            _watcher.Deleted += OnPluginDirChanged;
            _watcher.Renamed += OnPluginDirChanged;

            _watcher.EnableRaisingEvents = true;
            _logger.LogInformation("Plugin hot-reload watcher enabled for: {Path}", path);
        }

        private void OnPluginDirChanged(object sender, FileSystemEventArgs e)
        {
            var ext = Path.GetExtension(e.FullPath).ToLower();
            if (ext == ".json" || ext == ".exe" || ext == ".dll" || string.IsNullOrEmpty(ext))
            {
                _logger.LogDebug("Plugin file change detected: {File}, scheduling reload...", e.Name);
                _reloadTimer?.Stop();
                _reloadTimer?.Start();
            }
        }

        private async Task DoReloadAsync()
        {
            if (string.IsNullOrEmpty(_lastScanDir)) return;
            _logger.LogInformation("Hot-reloading plugins due to file changes...");
            await FullScanAsync(_lastScanDir);
        }

        private async Task FullScanAsync(string baseDir)
        {
            List<IPlugin> oldPlugins;
            lock (_lock)
            {
                oldPlugins = new List<IPlugin>(_plugins);
                _plugins.Clear();
                _eventHandlers.Clear();
                _skills.Clear();
            }

            foreach (var p in oldPlugins)
            {
                try { await p.StopAsync(); } catch { }
            }

            await PerformScanInternalAsync(baseDir);
        }

        private async Task PerformScanInternalAsync(string baseDir)
        {
            if (!Directory.Exists(baseDir))
            {
                _logger.LogWarning("Plugin directory not found: {BaseDir}", baseDir);
                return;
            }

            _logger.LogInformation("Scanning for plugins in: {BaseDir}", baseDir);
            var configFiles = Directory.GetFiles(baseDir, "plugin.json", SearchOption.AllDirectories);

            foreach (var configFile in configFiles)
            {
                try
                {
                    var json = await File.ReadAllTextAsync(configFile);
                    var config = JsonSerializer.Deserialize<PluginConfig>(json, new JsonSerializerOptions { PropertyNameCaseInsensitive = true });
                    
                    if (config == null || string.IsNullOrEmpty(config.Name)) continue;

                    IPlugin? plugin = null;
                    var pluginDir = Path.GetDirectoryName(configFile)!;

                    switch (config.Type?.ToLower())
                    {
                        case "process":
                            if (!string.IsNullOrEmpty(config.Executable))
                            {
                                var exePath = Path.Combine(pluginDir, config.Executable);
                                plugin = new ProcessPlugin(config, exePath, _logger, pluginDir);
                            }
                            break;
                        default:
                            _logger.LogWarning("Unsupported plugin type '{Type}' in {File}", config.Type, configFile);
                            break;
                    }

                    if (plugin != null)
                    {
                        await LoadPluginAsync(plugin);
                    }
                }
                catch (Exception ex)
                {
                    _logger.LogError(ex, "Error loading plugin from {File}", configFile);
                }
            }
        }

        public async Task RegisterSkillAsync(SkillCapability capability, Func<IPluginContext, string[], Task<string>> handler)
        {
            var pluginId = _currentLoadingPlugin?.Metadata.Id ?? "system";

            // 如果能力中没有定义意图，且当前正在加载插件，尝试从插件中获取
            if ((capability.Intents == null || capability.Intents.Count == 0) && _currentLoadingPlugin != null)
            {
                capability.Intents = _currentLoadingPlugin.Intents;
            }

            _skills.Add(new Skill { PluginId = pluginId, Capability = capability, Handler = handler });
            
            // 提取命令关键词并注册到 BotCmd
            var keywords = new HashSet<string>(StringComparer.OrdinalIgnoreCase);
            
            if (capability.Commands != null)
            {
                foreach (var cmd in capability.Commands)
                {
                    var k = cmd.Split(' ')[0].Trim();
                    if (!string.IsNullOrEmpty(k)) keywords.Add(k);
                }
            }

            if (capability.Intents != null)
            {
                foreach (var intent in capability.Intents)
                {
                    if (intent.Keywords != null)
                    {
                        foreach (var k in intent.Keywords)
                        {
                            if (!string.IsNullOrEmpty(k)) keywords.Add(k.Trim());
                        }
                    }
                }
            }

            if (keywords.Count > 0)
            {
                _serviceProvider.GetRequiredService<IBotCmdService>().RegisterExtraCommands(keywords);
            }

            _logger.LogInformation("Skill registered: {Name} (Plugin: {PluginId}, Commands: {Commands}, Keywords: {Keywords})", 
                capability.Name, pluginId, string.Join(", ", capability.Commands ?? Array.Empty<string>()), string.Join(", ", keywords));
            await Task.CompletedTask;
        }

        public async Task RegisterEventAsync(string eventType, Func<IPluginContext, Task> handler)
        {
            if (!_eventHandlers.ContainsKey(eventType))
            {
                _eventHandlers[eventType] = new List<Func<IPluginContext, Task>>();
            }
            _eventHandlers[eventType].Add(handler);
            await Task.CompletedTask;
        }

        public async Task LoadPluginAsync(IPlugin plugin)
        {
            try
            {
                var metadata = plugin.Metadata;
                _logger.LogInformation("Loading plugin: {Name} v{Version} by {Author}", metadata.Name, metadata.Version, metadata.Author);
                _plugins.Add(plugin);
                
                _currentLoadingPlugin = plugin;
                try
                {
                    await plugin.InitAsync(this);
                }
                finally
                {
                    _currentLoadingPlugin = null;
                }
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Failed to load plugin: {PluginName}", plugin.GetType().Name);
            }
        }

        public async Task<string> HandleEventAsync(EventBase ev, Func<string, Task>? replyDelegate = null)
        {
            // 1. 异步加载上下文数�?
            var userIdStr = ev.UserId;
            var groupIdStr = ev.GroupId;
            
            long userId = 0;
            long.TryParse(userIdStr, out userId);
            
            long groupId = 0;
            if (!string.IsNullOrEmpty(groupIdStr)) long.TryParse(groupIdStr, out groupId);
            
            var botId = ev.SelfId;

            var userTask = userId != 0 ? UserInfo.GetSingleAsync(userId) : Task.FromResult<UserInfo?>(null);
            var botTask = BotInfo.GetSingleAsync(botId);
            var groupTask = groupId != 0 ? GroupInfo.GetSingleAsync(groupId) : Task.FromResult<GroupInfo?>(null);
            var memberTask = (groupId != 0 && userId != 0) ? GroupMember.GetSingleAsync(groupId, userId) : Task.FromResult<GroupMember?>(null);

            await Task.WhenAll(userTask, botTask, groupTask, memberTask);

            var ctx = new PluginContext(
                ev, 
                "onebot", 
                botId.ToString(),
                _aiService,
                _i18nService,
                _logger,
                await userTask,
                await groupTask,
                await memberTask,
                await botTask,
                replyDelegate,
                musicReplyDelegate: null) // 这里暂时不传，因为 HandleEventAsync 通常不处理音乐
            {
                RawMessage = ev.RawMessage
            };

            return await DispatchAsync(ctx);
        }

        public async Task<string> CallSkillAsync(string skillName, IPluginContext ctx, string[] args)
        {
            var skill = _skills.FirstOrDefault(s => s.Capability.Name.Equals(skillName, StringComparison.OrdinalIgnoreCase) || 
                                                    s.Capability.Commands.Contains(skillName, StringComparer.OrdinalIgnoreCase));
            if (skill != null)
            {
                try
                {
                    return await skill.Handler(ctx, args);
                }
                catch (Exception ex)
                {
                    _logger.LogError(ex, "Error calling skill {SkillName}", skillName);
                    return $"Error: {ex.Message}";
                }
            }
            return $"Skill {skillName} not found";
        }

        public async Task<string> DispatchAsync(IPluginContext ctx)
        {
            _logger.LogInformation("[Dispatch] EventType: {EventType}, UserId: {UserId}, GroupId: {GroupId}, Message: '{Message}', SkillCount: {SkillCount}", 
                ctx.EventType, ctx.UserId, ctx.GroupId, ctx.RawMessage, _skills.Count);

            // 1. 处理通用事件分发
            if (_eventHandlers.TryGetValue(ctx.EventType, out var handlers))
            {
                _logger.LogInformation("[Dispatch] Found {Count} event handlers for {EventType}", handlers.Count, ctx.EventType);
                foreach (var handler in handlers)
                {
                    try 
                    { 
                        await handler(ctx); 
                    } 
                    catch (Exception ex) 
                    { 
                        _logger.LogError(ex, "[Dispatch] Error in event handler for {EventType}", ctx.EventType); 
                    }
                }
            }

            // 2. 处理消息指令 (仅限 PostType 为 message 的情况)
            if (ctx.IsMessage)
            {
                var message = ctx.RawMessage.Trim();
                if (string.IsNullOrEmpty(message)) return string.Empty;

                // 2.1 会话拦截逻辑
                var session = await _sessionManager.GetSessionAsync(ctx.UserId, ctx.GroupId);
                if (session != null)
                {
                    _logger.LogInformation("[Dispatch] Active session found for user {UserId}, Plugin: {PluginId}, Action: {Action}", ctx.UserId, session.PluginId, session.Action);
                    if (message == "取消")
                    {
                        await _sessionManager.ClearSessionAsync(ctx.UserId, ctx.GroupId);
                        return "✅ 已取消当前操作。";
                    }

                    // 填充会话信息到 Context
                    ctx.SessionAction = session.Action;
                    ctx.SessionStep = session.Step;
                    ctx.SessionData = session.DataJson;

                    if (!string.IsNullOrEmpty(session.ConfirmationCode))
                    {
                        if (message == session.ConfirmationCode)
                        {
                            _logger.LogInformation("Session confirmation code matched for {UserId}", ctx.UserId);
                            // 验证码匹配，允许继续执行，并标记为已确认
                            ctx.IsConfirmed = true;
                            await _sessionManager.ClearSessionAsync(ctx.UserId, ctx.GroupId);
                            // 继续向下执行技能匹配
                        }
                        else
                        {
                            return $"⚠️ 当前有待确认的操作：{session.Action}\n请输入验证码【{session.ConfirmationCode}】以确认，或发送“取消”退出。";
                        }
                    }
                    else
                    {
                        // 2.1.2 通用对话模式 (多步对话)
                        // 查找发起该会话的插件所注册的技能
                        var targetSkill = _skills.FirstOrDefault(s => s.PluginId == session.PluginId);
                        if (targetSkill != null)
                        {
                            try
                            {
                                _logger.LogInformation("[Dispatch] Message handled by session for user {UserId}, Plugin: {PluginId}", ctx.UserId, session.PluginId);
                                return await targetSkill.Handler(ctx, Array.Empty<string>());
                            }
                            catch (Exception ex)
                            {
                                _logger.LogError(ex, "Error in multi-step session for plugin {PluginId}", session.PluginId);
                                return $"对话处理错误: {ex.Message}";
                            }
                        }
                    }
                }

                // 2.2 直接指令匹配 (命令前缀优先)
                foreach (var skill in _skills)
                {
                    foreach (var cmd in skill.Capability.Commands)
                    {
                        // 提取命令的核心关键词（第一个空格前的部分）
                        var cmdKey = cmd.Split(' ')[0].Trim();
                        if (string.IsNullOrEmpty(cmdKey)) continue;

                        if (message.StartsWith(cmdKey, StringComparison.OrdinalIgnoreCase))
                        {
                            var args = message.Substring(cmdKey.Length).Trim().Split(new[] { ' ' }, StringSplitOptions.RemoveEmptyEntries);
                            _logger.LogInformation("[Dispatch] Command match: Skill '{SkillName}' matched command keyword '{CommandKey}' (from '{FullCommand}')", 
                                skill.Capability.Name, cmdKey, cmd);
                            var result = await skill.Handler(ctx, args);
                            return result;
                        }
                    }
                }

                // 2.3 意图识别匹配 (Regex & Keywords)
                foreach (var skill in _skills)
                {
                    foreach (var intent in skill.Capability.Intents)
                    {
                        bool isMatch = false;

                        // 2.3.1 正则匹配
                        if (!string.IsNullOrEmpty(intent.Regex))
                        {
                            try
                            {
                                if (System.Text.RegularExpressions.Regex.IsMatch(message, intent.Regex, System.Text.RegularExpressions.RegexOptions.IgnoreCase))
                                {
                                    isMatch = true;
                                }
                            }
                            catch (Exception ex)
                            {
                                _logger.LogWarning(ex, "Invalid regex in intent {IntentName} for skill {SkillName}", intent.Name, skill.Capability.Name);
                            }
                        }

                        // 2.3.2 关键词匹配
                        if (!isMatch && intent.Keywords != null && intent.Keywords.Length > 0)
                        {
                            if (intent.Keywords.Any(k => message.Contains(k, StringComparison.OrdinalIgnoreCase)))
                            {
                                isMatch = true;
                            }
                        }

                        if (isMatch)
                        {
                            _logger.LogInformation("Intent matched: {IntentName} -> Skill: {SkillName}", intent.Name, skill.Capability.Name);
                            // 意图匹配通常将整条消息视为参数，或由插件自行解析 Context.RawMessage
                            return await skill.Handler(ctx, Array.Empty<string>());
                        }
                    }
                }

                // 2.4 AI 兜底 (AI Fallback)
                _logger.LogInformation("[Dispatch] No explicit match found, falling back to AI for: {Message}", message);
                var aiResult = await _aiService.ChatWithContextAsync(message, ctx);
                _logger.LogInformation("[Dispatch] AI result: {Result}", aiResult);
                return aiResult;
            }

            return string.Empty;
        }
    }
}


