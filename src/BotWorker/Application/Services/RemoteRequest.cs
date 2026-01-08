using System.Collections.Concurrent;
using Newtonsoft.Json;
using BotWorker.Bots.BotMessages;

namespace BotWorker.Core.Services
{
    public class RemoteRequest
    {
        private readonly ConcurrentDictionary<string, PendingRequest> _pendingRequests = new();

        public string Register(string methodName, BotMessage context,  object[] args)
        {
            string requestId = Guid.NewGuid().ToString("N");
            var pending = new PendingRequest(requestId, methodName, context, args);
            _pendingRequests[requestId] = pending;
            return requestId;
        }

        public bool Complete(string requestId, string resultJson)
        {
            if (_pendingRequests.TryRemove(requestId, out var pending))
            {
                Task.Run(async () =>
                {
                    pending.Context.Answer = JsonConvert.DeserializeObject<string>(resultJson) ?? "";
                    pending.Context.IsSend = true;
                    //pending.Context.RecallAfterMs = 10000;
                    await pending.Context.SendMessageAsync();
                });

                return true;
            }

            return false;
        }

        public int PendingCount => _pendingRequests.Count;
    }

}


