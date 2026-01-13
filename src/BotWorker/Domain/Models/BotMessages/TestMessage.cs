namespace BotWorker.Domain.Models.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        public async Task<string> GetTestItAsync()        
        {
            return await Task.FromResult("");
        }
    }
}
