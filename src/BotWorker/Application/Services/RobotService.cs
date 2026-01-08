using BotWorker.Domain.Interfaces;

namespace BotWorker.Application.Services
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


