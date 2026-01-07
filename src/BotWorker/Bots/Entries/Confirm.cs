using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.Entries
{
    public class Confirm : MetaData<Confirm>
    {
        public override string TableName => "Confirm";
        public override string KeyField => "Id";
    }
}
