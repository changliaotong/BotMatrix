using BotWorker.Bots.BotMessages;
using BotWorker.Bots.Interfaces;

namespace BotWorker.Bots.Services
{
    public class RobotService : IBotHandlerMessage, IBotSendMessage
    {
        public async Task HandleBotMessageAsync(BotMessage context)
        {         
            ArgumentNullException.ThrowIfNull(context);
            await context.HandleEventAsync();            
        }

        public async Task SendFinalMessageAsync(BotMessage context)
        {
            ArgumentNullException.ThrowIfNull(context);
            await context.SendMessageAsync();
        }
    }      
}
