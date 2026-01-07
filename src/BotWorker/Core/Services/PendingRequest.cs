using sz84.Bots.BotMessages;

namespace sz84.Core.Services
{
    public class PendingRequest(string requestId, string methodName, BotMessage context, object[] args)
    {
        public string RequestId { get; } = requestId;
        public BotMessage Context { get; } = context;
        public string MethodName { get; } = methodName;
        public object[] Args { get; } = args;
    }

}
