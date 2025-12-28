using System;
using BotMatrix.SDK;

namespace DemoPlugin
{
    class Program
    {
        static void Main(string[] args)
        {
            var plugin = new BotMatrixPlugin();

            plugin.OnMessage(ctx =>
            {
                var text = ctx.Event.Payload["text"]?.ToString() ?? "";
                if (text.StartsWith("/csharp "))
                {
                    var content = text.Substring("/csharp ".Length);
                    ctx.Reply($"C# SDK Echo: {content}");
                }
            });

            plugin.Run();
        }
    }
}
