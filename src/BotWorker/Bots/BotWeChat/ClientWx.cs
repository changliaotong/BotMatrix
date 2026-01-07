using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.BotWeChat
{
    public class ClientWx : MetaData<ClientWx>
    {
        public override string TableName => "wx_client";
        public override string KeyField => "client_qq";
    }
}
