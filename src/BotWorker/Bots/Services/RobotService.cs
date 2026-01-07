using sz84.Bots.BotMessages;
using sz84.Bots.Interfaces;

namespace sz84.Bots.Services
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
