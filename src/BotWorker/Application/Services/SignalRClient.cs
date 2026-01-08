using Microsoft.AspNetCore.SignalR.Client;
using BotWorker.Bots.Entries;

namespace BotWorker.Core.Services
{
    public class SignalRClient
    {
        public HubConnection _conn;
        private readonly List<string> _serverUrls;
        private int _currentServerIndex = 0;
        private readonly List<PendingSignalRMessage> _queue = [];
        private readonly object _lock = new();
        private readonly TimeSpan _maxAge = TimeSpan.FromMinutes(5);
        private readonly Action<string>? _log;
        private bool _isReconnecting;
        private readonly long _userId;
        private readonly SemaphoreSlim _connectLock = new(1, 1);
        private bool _isManualStop = false;
        private readonly Dictionary<string, Delegate> _handlers = [];
        private readonly List<(string method, Delegate handler)> _pendingRegistrations = [];
        private readonly HashSet<string> _registered = [];
        private Timer? _primaryCheckTimer;

        private async Task SwitchToBackupAsync() => await SwitchToServerAsync((_currentServerIndex + 1) % _serverUrls.Count);
        
        private async Task SwitchToPrimaryAsync() => await SwitchToServerAsync(0);

        private void StartPrimaryCheckTimer()
        {
            _primaryCheckTimer = new Timer(async _ =>
            {
                if (_currentServerIndex == 0) return; 

                var primaryUrl = _serverUrls[0];
                var testConn = new HubConnectionBuilder()
                    .WithUrl(primaryUrl)
                    .WithAutomaticReconnect()
                    .Build();

                try
                {
                    await testConn.StartAsync();
                    await testConn.StopAsync();

                    InfoMessage("[��������] ���������ָ����л�����������...");
                    await SwitchToPrimaryAsync(); 
                }
                catch
                {
                    InfoMessage("[��������] ���������Բ�����...");
                }
            }, null, TimeSpan.FromSeconds(15), TimeSpan.FromSeconds(30)); // ÿ 30 ����һ��
        }

        private async Task SwitchToServerAsync(int index)
        {
            if (index == _currentServerIndex) return;

            try { await _conn.StopAsync(); } catch { }

            _currentServerIndex = index;
            var url = _serverUrls[_currentServerIndex];

            _conn = CreateConnection(_userId, _serverUrls[_currentServerIndex]);
            AttachConnectionEvents();

            InfoMessage($"[SignalR] ������{(_currentServerIndex == 0 ? "��" : "����")}��������{url}");

            if (_currentServerIndex != 0)
                StartPrimaryCheckTimer(); 
            else
                _primaryCheckTimer?.Dispose(); 
        }


        public async Task StopManuallyAsync()
        {
            _isManualStop = true;
            try
            {
                await _conn.StopAsync();
            }
            finally
            {
                _isManualStop = false;
            }
        }

        public async Task CancelStream(string msgGuid)
        {
            if (_conn != null)
            {
                await SendMessageAsync("CancelStream", msgGuid.ToString());
            }
        }

        public SignalRClient(long userId, List<string> serverUrls, Action<string>? log = null)
        {
            if (serverUrls == null || serverUrls.Count == 0)
                throw new ArgumentException("�����ṩһ����������ַ");

            _userId = userId;
            _serverUrls = serverUrls;
            _log = log;
            _conn = CreateConnection(userId, _serverUrls[_currentServerIndex]);
            AttachConnectionEvents();
        }

        #region ע����Ϣ�¼�

        public void On<T1>(string method, Action<T1> handler)
        {
            _pendingRegistrations.Add((method, handler));
            if (_conn != null && _conn.State == HubConnectionState.Connected)
            {
                _conn.Remove(method);
                _conn.On<T1>(method, handler);
                _registered.Add(method);
            }
        }

        public void On<T1, T2>(string method, Action<T1, T2> handler)
        {
            _pendingRegistrations.Add((method, handler));
            if (_conn != null && _conn.State == HubConnectionState.Connected)
            {
                _conn.Remove(method);
                _conn.On<T1, T2>(method, handler);
                _registered.Add(method);
            }
        }

        public void On<T1, T2, T3>(string method, Action<T1, T2, T3> handler)
        {
            _pendingRegistrations.Add((method, handler));
            if (_conn != null && _conn.State == HubConnectionState.Connected)
            {
                _conn.Remove(method);
                _conn.On<T1, T2, T3>(method, handler);
                _registered.Add(method);
            }
        }

        public void On<T1, T2, T3, T4>(string method, Action<T1, T2, T3, T4> handler)
        {
            _pendingRegistrations.Add((method, handler));
            if (_conn != null && _conn.State == HubConnectionState.Connected)
            {
                _conn.Remove(method);
                _conn.On<T1, T2, T3, T4>(method, handler);
                _registered.Add(method);
            }
        }

        public void RegisterHandler<T1>(string methodName, Action<T1> handler)
        {
            _conn?.Remove(methodName);
            _conn?.On<T1>(methodName, handler);
            _handlers[methodName] = handler;
        }

        public void RegisterHandler<T1, T2>(string methodName, Action<T1, T2> handler)
        {
            _conn?.Remove(methodName);
            _conn?.On<T1, T2>(methodName, handler);
            _handlers[methodName] = handler;
        }

        public void RegisterHandler<T1, T2, T3>(string methodName, Action<T1, T2, T3> handler)
        {
            if (_conn == null) throw new InvalidOperationException("[SignalR] connection is not started.");
            _conn.Remove(methodName);
            _conn.On<T1, T2, T3>(methodName, handler);
            _handlers[methodName] = handler;
        }

        public void RegisterHandler<T1, T2, T3, T4>(string methodName, Action<T1, T2, T3, T4> handler)
        {
            if (_conn == null) throw new InvalidOperationException("[SignalR] connection is not started.");
            _conn.Remove(methodName);
            _conn.On<T1, T2, T3, T4>(methodName, handler);
            _handlers[methodName] = handler;
        }

        private void RegisterHandlerDynamic(string methodName, Delegate handler)
        {
            var methodInfo = typeof(HubConnection).GetMethods()
                .Where(m => m.Name == "On" && m.IsGenericMethod)
                .Select(m => new
                {
                    Method = m,
                    Params = m.GetParameters(),
                    GenArgs = m.GetGenericArguments()
                })
                .FirstOrDefault(m => m.Params.Length == 2
                    && m.Params[0].ParameterType == typeof(string)
                    && m.Params[1].ParameterType.Name.StartsWith("Action`"));

            if (methodInfo == null)
                throw new Exception("[SignalR] �Ҳ������ʵ� On ����");

            var genericTypes = handler.Method.GetParameters().Select(p => p.ParameterType).ToArray();

            var genericMethod = methodInfo.Method.MakeGenericMethod(genericTypes);
            genericMethod.Invoke(_conn, new object[] { methodName, handler });
        }

        private void RegisterAllHandlers()
        {
            if (_conn == null) return;

            foreach (var kvp in _handlers)
            {
                _conn.Remove(kvp.Key);
                RegisterHandlerDynamic(kvp.Key, kvp.Value);
            }
        }

        #endregion ��Ϣ�¼�

        // ������� StartAsync() �ɹ������
        private void RegisterMessageHandlers()
        {
            On<string, string, string>("ReceiveRequestMessage", (requestId, method, args) =>
              OnRequestMessageReceived?.Invoke(requestId, method, args));

            On<string, long, string>("ReceiveMessage", (guid, selfId, message) =>
                OnMessageReceived?.Invoke(guid, selfId, message));

            On<string, string>("ReceiveBotMessage", (guid, message) =>
            {
                OnBotMessageReceived?.Invoke(guid, message);
            });

            On<string, string>("ReceiveProxyMessage", (guid, message) =>
                OnProxyMessageReceived?.Invoke(guid, message));

            On<string, long, long, long>("ReceiveMentionMessage", (guid, gid, oid, bid) =>
                OnMentionMessageReceived?.Invoke(guid, gid, oid, bid));

            On<string>("ReceiveStreamBeginMessage", guid =>
                OnStreamBeginMessageReceived?.Invoke(guid));

            On<string, string>("ReceiveStreamMessage", (guid, message) =>
                OnStreamMessageReceived?.Invoke(guid, message));

            On<string>("ReceiveStreamEndMessage", guid =>
                OnStreamEndMessageReceived?.Invoke(guid));
        }

        #region ��Ϣ�¼����壨���ⲿ���ģ�

        public event Action<string, string, string>? OnRequestMessageReceived;
        public event Action<string>? OnStreamEndMessageReceived;
        public event Action<string>? OnStreamBeginMessageReceived;
        public event Action<string, string>? OnStreamMessageReceived;
        public event Action<string, long, string>? OnMessageReceived;

        public event Action<string, string>? OnBotMessageReceived;
        public event Action<string, string>? OnProxyMessageReceived;
        public event Action<string, long, long, long>? OnMentionMessageReceived;
        #endregion

        public async Task<bool> SendResponseMessage(string guid, string message)
            => await SendMessageAsync<bool>("SendResponse", guid, message);

        public async Task<bool> SendMessage(string guid, string message)
            => await SendMessageAsync<bool>("SendMessage", guid, message);

        public async Task<bool> SendBotMessage(string guid, string message)
            => await SendMessageAsync<bool>("SendBotMessage", guid, message);

        public async Task<bool> SendMentionMessage(string guid, string groupOpenoid, long officialBot)
            => await SendMessageAsync<bool>("SendMentionMessage", guid, groupOpenoid, officialBot);

        public async Task<bool> BroadCastAsync(string guid, string message)
            => await SendMessageAsync<bool>("BroadCastMessage", guid, message);

        public async Task SendStreamUserMessage(string messsage)
            => await SendMessageAsync("SendStreamUserMessage", messsage);

        public async Task SendProxyMessage(string userId, string guid, string messsage)
            => await SendMessageAsync("SendProxyMessage", userId, guid, messsage);

        private async Task<T?> SendMessageAsync<T>(string method, params object[] args)
        {
            try
            {
                using var cts = new CancellationTokenSource(120000);
                T result = await _conn.InvokeCoreAsync<T>(method, args, cancellationToken: cts.Token);
                return result;
            }
            catch (Exception ex)
            {
                InfoMessage($"[SignalR] ����ʧ�ܣ�{ex.GetBaseException()}");
            }

            CacheMessage(method, args);

            return default;
        }

        private async Task SendMessageAsync(string method, params object?[] args)
        {
            try
            {
                using var cts = new CancellationTokenSource(120000);
                await _conn.InvokeCoreAsync(method, args, cancellationToken: cts.Token);
                return;
            }
            catch (Exception ex)
            {
                InfoMessage($"[SignalR] ����ʧ�ܣ�{ex.GetBaseException()}");
            }
        }

        private static HubConnection CreateConnection(long userId, string url)
        {
            return new HubConnectionBuilder()
                .WithUrl(url, options =>
                {
                    options.Headers.Add("userId", userId.ToString());
                })
                .WithAutomaticReconnect()
                .Build();
        }

        private void AttachConnectionEvents()
        {
            _conn.Closed += async (error) =>
            {
                if (_isManualStop)
                {
                    InfoMessage("[SignalR] �ֶ��Ͽ��������Զ�����");
                    return;
                }

                InfoMessage($"[SignalR] ���ӹرգ�{error?.Message}");
                await EnsureConnectedAsync();
            };

            _conn.Reconnected += async (_) =>
            {
                RegisterMessageHandlers();
                RegisterAllHandlers();

                InfoMessage("[SignalR] �������������ط�������Ϣ");
                await RetryAllAsync();
            };
        }

        public async Task StartAsync() => await EnsureConnectedAsync();

        private void CacheMessage(string method, object[] args)
        {
            lock (_lock)
            {
                _queue.Add(new PendingSignalRMessage
                {
                    MethodName = method,
                    Arguments = args,
                    CreatedAt = DateTime.UtcNow
                });
            }
        }

        public async Task RetryAllAsync()
        {
            if (_conn.State != HubConnectionState.Connected)
            {
                InfoMessage("[SignalR] ��ǰδ���ӣ���������");
                return;
            }

            List<PendingSignalRMessage> toRetry;
            var now = DateTime.UtcNow;

            lock (_lock)
            {
                toRetry = [.. _queue.Where(m => now - m.CreatedAt < _maxAge)];
            }

            foreach (var msg in toRetry)
            {
                try
                {
                    using var cts = new CancellationTokenSource(30000);
                    await Task.Delay(500);
                    await _conn.InvokeCoreAsync(msg.MethodName, msg.Arguments, cancellationToken: cts.Token);
                    lock (_lock) { _queue.Remove(msg); }
                    InfoMessage($"[SignalR] ���Գɹ���{msg.MethodName}");
                }
                catch (Exception ex)
                {
                    InfoMessage($"[SignalR] ����ʧ�ܣ�{ex.Message}");
                }
            }

            lock (_lock)
                _queue.RemoveAll(m => now - m.CreatedAt >= _maxAge);
        }

        public async Task EnsureConnectedAsync()
        {
            if (_isReconnecting) return;
            _isReconnecting = true;

            int[] retryDelays = [4, 8, 16, 32, 64, 128, 256, 356, 456, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536];

            try
            {
                await RetryHelper.RetryAsync(async () =>
                {
                    if (_conn.State == HubConnectionState.Connected)
                    {
                        InfoMessage("[SignalR] ��ǰ�����ӣ����� Stop");
                        return;
                    }

                    if (_conn.State == HubConnectionState.Connecting || _conn.State == HubConnectionState.Reconnecting)
                    {
                        InfoMessage("[SignalR] ��ǰ���������У��ȴ��������");
                        return;
                    }

                    await _connectLock.WaitAsync();
                    try
                    {
                        if (_conn.State != HubConnectionState.Disconnected)
                        {
                            InfoMessage("[SignalR] ����ֹͣ����");
                            await StopManuallyAsync();
                        }
                    }
                    catch (Exception ex)
                    {
                        ErrorMessage($"[SignalR] Stop ʧ�ܣ�{ex.Message}");
                    }
                    finally
                    {
                        _connectLock.Release();
                    }    

                    InfoMessage($"[SignalR] ��������{(_currentServerIndex == 0 ? "��" : "����")}��������{_serverUrls[_currentServerIndex]}");

                    _conn = CreateConnection(_userId, _serverUrls[_currentServerIndex]);
                    AttachConnectionEvents();

                    await _conn.StartAsync();

                    RegisterMessageHandlers();
                    RegisterAllHandlers();

                    if (_conn.State != HubConnectionState.Connected)
                        throw new InvalidOperationException("����ʧ��");
                }, retryDelays, async delay =>
                {
                    InfoMessage($"[SignalR] ����ʧ�ܣ�{delay} �������...");
                    await SwitchToBackupAsync();
                });

                InfoMessage("[SignalR] ���ӳɹ�");
            }
            finally
            {
                _isReconnecting = false;
            }
        }
    }
}


