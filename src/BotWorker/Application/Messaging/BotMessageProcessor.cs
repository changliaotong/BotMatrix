namespace BotWorker.Application.Messaging
{
    public class BotMessageProcessor
    {
        private readonly BotCommandHandler _handler;
        public BotMessageProcessor(BotCommandHandler handler) => _handler = handler;
    }
}


