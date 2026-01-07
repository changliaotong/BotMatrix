using Newtonsoft.Json;
using System;
using System.Reflection;
using Newtonsoft.Json.Linq;

namespace sz84.Core.Services
{
    public static class ClientRequestHandler
    {
        // 注册的通用处理表：方法名 => 处理函数
        public static readonly Dictionary<string, (Func<object[], Task<object>> func, Type[] paramTypes)> Handlers = [];
            
        // 用于主程序调用处理服务端请求
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
                Console.WriteLine($"[处理异常] {methodName}: {ex.Message}");
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
