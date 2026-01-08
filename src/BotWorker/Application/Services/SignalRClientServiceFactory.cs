using Microsoft.AspNetCore.Components;
using Microsoft.Extensions.Configuration;

namespace BotWorker.Core.Services
{
    public static class SignalRClientServiceFactory
    {
        private static string HubUrlBackup = "http://192.168.0.69/ChatHub";

        public static SignalRClient CreateFromConfiguration(long userId, IConfiguration config)
        {
            var hubUrl = config["SignalR:HubUrl"];
            if (string.IsNullOrWhiteSpace(hubUrl))
                throw new InvalidOperationException("������ȱ�� SignalR:HubUrl");
            var servers = new List<string>
                {
                    hubUrl,
                    HubUrlBackup,
                };
            InfoMessage($"[SignalR] ���� SignalRClientService, HubUrl:{hubUrl}");
            return new SignalRClient(userId, servers);
        }

        public static SignalRClient CreateFromNavigation(long userId, NavigationManager nav)
        {
            var hubUrl = nav.ToAbsoluteUri("/ChatHub").ToString();
            var servers = new List<string>
                {
                    hubUrl,
                    HubUrlBackup,
                };
            return new SignalRClient(userId, servers);
        }

#if BLAZOR
            public static SignalRClientService CreateFromNavigation(NavigationManager nav)
            {
                var hubUrl = nav.ToAbsoluteUri("/ChatHub").ToString();
                return new SignalRClientService(hubUrl);
            }
#endif
    }

}


