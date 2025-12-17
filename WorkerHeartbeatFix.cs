// Worker端心跳响应修复代码
// 在 OneBotWebSocketClient 类的构造函数中，_client.MessageReceived.Subscribe 部分添加：

_client.MessageReceived.Subscribe(msg => 
{ 
    if (string.IsNullOrEmpty(msg.Text)) return; 

    try 
    { 
        var obj = JObject.Parse(msg.Text); 

        // =================== 新增：BotNexus心跳请求响应 ===================
        if (obj["action"]?.ToString() == "heartbeat")
        {
            var echo = obj["echo"]?.ToString();
            var workerId = obj["params"]?["worker_id"]?.ToString();
            
            if (!string.IsNullOrEmpty(echo))
            {
                // 发送心跳响应
                var response = new JObject
                {
                    ["status"] = "ok",
                    ["retcode"] = 0,
                    ["data"] = new JObject
                    {
                        ["heartbeat"] = "received",
                        ["timestamp"] = DateTimeOffset.UtcNow.ToUnixTimeSeconds()
                    },
                    ["echo"] = echo
                };
                
                // 发送响应
                _client.Send(response.ToString());
                Console.WriteLine($"[Heartbeat] Responded to heartbeat request: {echo}");
            }
            return; // 不继续处理心跳消息
        }
        // =================== 心跳响应结束 ===================

        // 原有的echo请求响应处理
        if (obj.TryGetValue("echo", out var echo) && _queues.TryGetValue(echo.ToString(), out var tuple)) 
        { 
            _queues[echo.ToString()] = (tuple.Item1, obj); 
            tuple.Item1.Release(); 
            return; // 处理完echo响应后返回
        } 

        // 原有的心跳消息处理（来自Bot的心跳）
        if (obj["post_type"]?.ToString() == "meta_event" && 
            obj["meta_event_type"]?.ToString() == "heartbeat") 
        { 
            // 这里处理的是Bot发送的心跳，不是我们需要响应的心跳请求
            // var status = obj["status"]; 
            // Console.WriteLine($"[Heartbeat] online={status?["online"]}, good={status?["good"]}"); 
            return;
        } 

        // 连接成功消息
        if (obj["post_type"]?.ToString() == "meta_event" && 
            obj["meta_event_type"]?.ToString() == "lifecycle" && 
            obj["sub_type"]?.ToString() == "connect") 
        { 
            Console.WriteLine("[Info] Connected to OneBot server"); 
            return;
        } 

        // API错误消息
        if (obj.ContainsKey("status") && obj["status"]?.ToString() == "failed") 
        { 
            Console.WriteLine($"[Error] {obj["message"]} (retcode: {obj["retcode"]})"); 
            return;
        } 

        // 普通事件
        if (obj.ContainsKey("self_id")) 
        { 
            var ev = EventBase.ParseRecv(obj); 
            if (ev != null) 
                EventRecv?.Invoke(this, ev); 
            return;
        } 

        Console.WriteLine($"[Unknown] {msg.Text}"); 
    } 
    catch (Exception ex) 
    { 
        Console.WriteLine($"[Exception] Failed to parse message: {ex.Message}"); 
    } 
});