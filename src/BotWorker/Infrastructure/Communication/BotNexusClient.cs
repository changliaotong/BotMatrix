using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Configuration;

namespace BotWorker.Infrastructure.Communication
{
    public class BotNexusClient : BackgroundService
    {
        private readonly ILogger<BotNexusClient> _logger;
        private readonly string _nexusUrl;
        private readonly string _workerId;
        private readonly string _platform;
        private ClientWebSocket? _webSocket;

        public BotNexusClient(
            ILogger<BotNexusClient> logger,
            IConfiguration configuration)
        {
            _logger = logger;
            var addr = configuration["nexus_addr"] ?? "ws://localhost:8081";
            if (!addr.EndsWith("/ws/workers"))
            {
                addr = addr.TrimEnd('/') + "/ws/workers";
            }
            _nexusUrl = addr;
            _workerId = configuration["worker_id"] ?? "csharp-worker";
            _platform = "C#";
        }

        protected override async Task ExecuteAsync(CancellationToken stoppingToken)
        {
            _logger.LogInformation("Starting BotNexus client, connecting to {NexusUrl}", _nexusUrl);

            while (!stoppingToken.IsCancellationRequested)
            {
                try
                {
                    _webSocket = new ClientWebSocket();
                    await ConnectAsync(stoppingToken);
                    await ReceiveLoopAsync(stoppingToken);
                }
                catch (Exception ex)
                {
                    _logger.LogError(ex, "BotNexus client error, retrying in 5 seconds...");
                    
                    if (_webSocket != null)
                    {
                        try { _webSocket.Dispose(); } catch { }
                        _webSocket = null;
                    }

                    await Task.Delay(5000, stoppingToken);
                }
            }
        }

        private async Task ConnectAsync(CancellationToken cancellationToken)
        {
            if (_webSocket == null) return;

            var uri = new Uri(_nexusUrl);
            _webSocket.Options.SetRequestHeader("X-Self-ID", _workerId);
            _webSocket.Options.SetRequestHeader("X-Platform", _platform);

            _logger.LogInformation("Connecting to BotNexus at {NexusUrl}...", _nexusUrl);
            await _webSocket.ConnectAsync(uri, cancellationToken);
            _logger.LogInformation("Connected to BotNexus!");

            // Send Lifecycle Event
            var lifecycleEvent = new
            {
                post_type = "meta_event",
                meta_event_type = "lifecycle",
                sub_type = "connect",
                self_id = _workerId,
                platform = _platform,
                time = DateTimeOffset.UtcNow.ToUnixTimeSeconds()
            };

            var json = JsonSerializer.Serialize(lifecycleEvent);
            var buffer = Encoding.UTF8.GetBytes(json);
            await _webSocket.SendAsync(new ArraySegment<byte>(buffer), WebSocketMessageType.Text, true, cancellationToken);
        }

        private async Task ReceiveLoopAsync(CancellationToken cancellationToken)
        {
            if (_webSocket == null) return;

            var buffer = new byte[1024 * 4];
            while (_webSocket.State == WebSocketState.Open && !cancellationToken.IsCancellationRequested)
            {
                var result = await _webSocket.ReceiveAsync(new ArraySegment<byte>(buffer), cancellationToken);
                if (result.MessageType == WebSocketMessageType.Close)
                {
                    await _webSocket.CloseAsync(WebSocketCloseStatus.NormalClosure, "Closing", cancellationToken);
                    _logger.LogInformation("BotNexus connection closed by server.");
                    break;
                }

                var message = Encoding.UTF8.GetString(buffer, 0, result.Count);
                _logger.LogDebug("Received message from BotNexus: {Message}", message);
                
                // Handle incoming commands if any
            }
        }
    }
}
