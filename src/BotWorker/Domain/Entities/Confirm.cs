using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities
{
    public class Confirm : MetaData<Confirm>
    {
        public override string TableName => "Confirm";
        public override string KeyField => "Id";
    }
}
