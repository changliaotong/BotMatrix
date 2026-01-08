using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Models.Messages.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        public string HandleFileMessage()
        {
            if (IsGuild)
            {
                //
            }
            else if (IsMirai)
            {

            }

            return $"";
        }
    }
}
