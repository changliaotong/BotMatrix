namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage
{
        public async Task<string> GetTestItAsync()        
        {
            return await Task.FromResult("");
        }
    }
}
