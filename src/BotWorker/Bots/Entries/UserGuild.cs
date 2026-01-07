using BotWorker.Bots.Users;
using BotWorker.Common.Exts;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.Entries
{
    public class UserGuild : MetaData<UserGuild>
    {
        public override string TableName => "User";
        public override string KeyField => "UserOpenid";

        public const long MIN_USER_ID = 980000000000;
        public const long MAX_USER_ID = 990000000000;

        public static long GetUserId(long botUin, string userOpenid, string groupOpenid)
        {
            if (userOpenid.IsNull()) 
                return 0;

            var userId = GetTargetUserId(userOpenid);
            if (userId != 0)
            {
                var bot = GetLong("isnull(BotUin, 0)", userOpenid);
                if (bot != botUin)
                    SetValue("BotUin", botUin, userOpenid);
                return userId;
            }

            userId = GetMaxUserId();
            int i = UserInfo.Append(botUin, 0, userId, "", 0, userOpenid, groupOpenid);
            return i == -1 ? 0 : userId;
        }

        public static long GetTargetUserId(string userOpenid)
        {
            return GetLong("isnull(TargetUserId, Id)", userOpenid);
        }

        private static long GetMaxUserId()
        {
            var userId = GetWhere("max(Id)", $"Id > {MIN_USER_ID} and Id < {MAX_USER_ID}").AsLong();
            return userId <= MIN_USER_ID ? MIN_USER_ID + 1 : userId + 1;
        }

        public static string GetUserOpenid(long selfId, long user)
        {
            return GetValueAandB<string>("UserOpenid", "TargetUserId", user, "BotUin", selfId);
        }

    }
}
