using System;
using System.Net.WebSockets;
using System.Text;
using System.Threading;
using System.Threading.Tasks;

namespace WxBotClient
{
    /// <summary>
    /// 一个优化的 WebSocket 客户端示例，展示了如何实现“快速重连”策略。
    /// 适用于服务端频繁重启或网络不稳定的场景。
    /// </summary>
    public class BetterOneBotClient
    {
        // 替换为你的 WebSocket 地址
        private const string WS_URL = "ws://192.168.0.167:3111";
        private ClientWebSocket _ws;
        private CancellationTokenSource _cts;

        public async Task StartAsync()
        {
            _cts = new CancellationTokenSource();

            while (!_cts.IsCancellationRequested)
            {
                try
                {
                    // 1. 尝试连接（包含重试逻辑）
                    await ConnectWithRetryAsync(_cts.Token);

                    // 2. 连接成功后进入接收循环
                    await ReceiveLoopAsync(_cts.Token);
                }
                catch (OperationCanceledException)
                {
                    break;
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"[MainLoop] Critical error: {ex.Message}");
                }

                // 如果接收循环异常退出，稍作休息再次进入连接流程
                if (!_cts.IsCancellationRequested)
                {
                    Console.WriteLine("[MainLoop] Session ended. Preparing to reconnect...");
                    await Task.Delay(1000, _cts.Token);
                }
            }
        }

        private async Task ConnectWithRetryAsync(CancellationToken token)
        {
            int retryCount = 0;
            
            while (!token.IsCancellationRequested)
            {
                try
                {
                    // 每次连接前必须重新实例化 ClientWebSocket
                    if (_ws != null)
                    {
                        try { _ws.Dispose(); } catch { }
                    }
                    _ws = new ClientWebSocket();

                    // 设置连接超时（关键点：防止 ConnectAsync 卡死太久）
                    // 建议设置为 2-3 秒，如果服务端重启很快，超时越短重试越快
                    using (var connectCts = CancellationTokenSource.CreateLinkedTokenSource(token))
                    {
                        connectCts.CancelAfter(TimeSpan.FromSeconds(3));
                        
                        Console.WriteLine($"[Connect] Connecting to {WS_URL}... (Attempt {retryCount + 1})");
                        await _ws.ConnectAsync(new Uri(WS_URL), connectCts.Token);
                    }

                    Console.WriteLine("[Connect] Connected successfully!");
                    return; // 连接成功，退出重试循环
                }
                catch (Exception ex)
                {
                    // 关键优化：快速重连策略
                    // 服务端重启通常只需要几秒钟。
                    // 策略：前 10 次失败仅等待 500ms，之后等待 5s。
                    // 这样当服务端在 4s 后恢复时，客户端能立刻连上，而不是等下一个 5s 或 10s 周期。
                    int delayMs = retryCount < 20 ? 500 : 5000;
                    
                    Console.WriteLine($"[Connect] Failed: {ex.Message}. Retrying in {delayMs}ms...");
                    
                    try
                    {
                        await Task.Delay(delayMs, token);
                    }
                    catch (OperationCanceledException) { break; }
                    
                    retryCount++;
                }
            }
        }

        private async Task ReceiveLoopAsync(CancellationToken token)
        {
            var buffer = new byte[1024 * 16];
            try
            {
                while (_ws.State == WebSocketState.Open && !token.IsCancellationRequested)
                {
                    var result = await _ws.ReceiveAsync(new ArraySegment<byte>(buffer), token);

                    if (result.MessageType == WebSocketMessageType.Close)
                    {
                        Console.WriteLine($"[Receive] Server closed connection: {result.CloseStatusDescription}");
                        await _ws.CloseOutputAsync(WebSocketCloseStatus.NormalClosure, "Ack", CancellationToken.None);
                        break;
                    }

                    // 处理消息...
                    string msg = Encoding.UTF8.GetString(buffer, 0, result.Count);
                    // Console.WriteLine($"[Receive] Msg len={result.Count}");
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[Receive] Error: {ex.Message}");
            }
            finally
            {
                Console.WriteLine("[Receive] Loop exited.");
            }
        }
        
        public void Stop()
        {
            _cts.Cancel();
        }
    }
}
