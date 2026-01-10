using Newtonsoft.Json;
using System;
using System.Reflection;
using Newtonsoft.Json.Linq;

namespace BotWorker.Application.Services
{
    public static class ClientRequestHandler
    {
        // ע���ͨ�ô����������� => �������
        public static readonly Dictionary<string, (Func<object[], Task<object>> func, Type[] paramTypes)> Handlers = [];
            
        // ������������ô�����������
        public static async Task<string> HandleRequest(string methodName, string args)
        {
            object? result = null;
            try
            {
                var (func, paramTypes) = Handlers[methodName];
                var jArray = JArray.Parse(args);
                var typedArgs = new object[paramTypes.Length];
                for (int i = 0; i < paramTypes.Length; i++)
                {
                    typedArgs[i] = jArray[i].ToObject(paramTypes[i])!;
                }
                result = await func(typedArgs);
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[�����쳣] {methodName}: {ex.Message}");
            }

            return JsonConvert.SerializeObject(result);
        }

        public class UserProfile
        {
            public string? Nickname { get; set; }
            public int Age { get; set; }
            public long QQ { get; set; }
        }
    }

}


