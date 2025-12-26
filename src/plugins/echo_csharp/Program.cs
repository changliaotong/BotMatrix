using System;
using System.IO;
using System.Text.Json;

namespace EchoCSharpPlugin
{
    public class EventMessage
    {
        public string id { get; set; }
        public string type { get; set; }
        public string name { get; set; }
        public object payload { get; set; }
    }

    public class Action
    {
        public string type { get; set; }
        public string target { get; set; }
        public string target_id { get; set; }
        public string text { get; set; }
    }

    public class ResponseMessage
    {
        public string id { get; set; }
        public bool ok { get; set; }
        public Action[] actions { get; set; }
    }

    class Program
    {
        static void Main(string[] args)
        {
            var stdin = Console.In;
            var stdout = Console.Out;
            
            while (true)
            {
                string line = stdin.ReadLine();
                if (line == null)
                    break;
                
                try
                {
                    var msg = JsonSerializer.Deserialize<EventMessage>(line);
                    
                    if (msg.type == "event" && msg.name == "on_message")
                    {
                        var payload = msg.payload as JsonElement;
                        string text = payload.GetProperty("text").GetString();
                        string target = payload.GetProperty("from").GetString();
                        string target_id = payload.GetProperty("group_id").GetString();
                        
                        var response = new ResponseMessage
                        {
                            id = msg.id,
                            ok = true,
                            actions = new[]
                            {
                                new Action
                                {
                                    type = "send_message",
                                    target = target,
                                    target_id = target_id,
                                    text = $"C# Echo: {text}"
                                }
                            }
                        };
                        
                        string json = JsonSerializer.Serialize(response);
                        stdout.WriteLine(json);
                        stdout.Flush();
                    }
                }
                catch (Exception ex)
                {
                    Console.Error.WriteLine($"Error: {ex.Message}");
                }
            }
        }
    }
}