using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Models.BotMessages
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
