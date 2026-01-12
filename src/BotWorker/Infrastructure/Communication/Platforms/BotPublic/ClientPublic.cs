namespace BotWorker.Infrastructure.Communication.Platforms.BotPublic
{
    public class ClientPublic : MetaData<ClientPublic>
    {
        public const long startUserId = 4104967295;

        public override string TableName => "PublicUser";
        public override string KeyField => "Id";

        public const string regexRec = @"^推荐[人]*(qq|q|号码)*[ :：　+＋十]*(?<recQQ>[1-9]\d{4,10})?";
        public const string format = "格式：推荐人+对方QQ\n例如：\n推荐人：{客服QQ}";

        //早喵AI
        public const string keyZaomiao = "gh_c9dfbd45d42f";
        // 畅聊通AI
        public const string keyCompany = "gh_2158fa6520a3";
        // 指路天使机器人
        public const string keyRobot = "gh_5696f9a0fae9";
        // 彭光辉
        public const string keyPenggh = "gh_f184bf294a46";

        // QQ是否已绑定某公众号
        public static bool ExistsQQ(string botKey, long userId)
        {
            return IsBind(userId) && ExistsAandB("BotKey", botKey, "UserId", userId);
        }

        // 是否关注了官方公众号 早喵AI、畅聊通AI、指路天使机器人、彭光辉
        public static bool SubscribeCompayPublic(long userId)
        {
            return IsBind(userId) && ExistsWhere($"BotKey in ({keyZaomiao.Quotes()}, {keyCompany.Quotes()}, {keyRobot.Quotes()}, {keyPenggh.Quotes()}) and UserId={userId}");
        }

        // 是否已经绑定的号码，即非自编号
        public static bool IsBind(long userId)
        {
            return userId < startUserId || userId > 90000000000;
        }

        public static bool IsNotBind(long userId)
        {
            return !IsBind(userId);
        }

        // 获得ID
        public static long GetId(string botKey, string clientKey)
        {
            string res = GetWhere("Id", $"BotKey = {botKey.Quotes()} and UserKey = {clientKey.Quotes()}", "Id desc");
            return res.AsLong();
        }

        // 通过机器人KEY与客户KEY获得对应的QQ号码，没有则添加。
        public static long GetUserId(string botKey, string clientKey)
        {
            return Exists(GetId(botKey, clientKey))
                ? GetLong("UserId", GetId(botKey, clientKey))
                : Append(botKey, clientKey) == -1 ? 0 : GetUserId(botKey, clientKey);
        }

        // TOKEN
        public static string GetBindToken(string botKey, string clientKey)
        {
            return (botKey + clientKey).MD5()[7..23];
        }

        //邀请码
        public static string InviteCode(string botKey, string UserKey)
        {
            string sql = $"select BindToken from {FullName} where BotKey = {botKey.Quotes()} and UserKey = {UserKey.Quotes()}";
            return QueryScalar<string>(sql) ?? "";
        }

        // 推荐人积分处理
        public static string GetRecRes(long botUin, long groupId, string groupName, long userId, string name, string botKey, string clientKey, string message)
        {
            return GetRecResAsync(botUin, groupId, groupName, userId, name, botKey, clientKey, message).GetAwaiter().GetResult();
        }

        public static async Task<string> GetRecResAsync(long botUin, long groupId, string groupName, long userId, string name, string botKey, string clientKey, string message)
        {
            if (!IsBind(userId))
                return "请先发送【领积分】完成积分任务";

            long clientOid = GetId(botKey, clientKey);
            string recUserIdStr = GetValue("RecUserId", clientOid);
            if (recUserIdStr != "")
                return $"推荐人已登记为：{recUserIdStr}\n{format}";

            string recUserIdInput = message.RegexGetValue(regexRec, "RecUserId");

            if (recUserIdInput.IsNull())
                return format;

            if (recUserIdInput == userId.ToString())
                return "推荐人不能是自己";

            long recUserId = recUserIdInput.AsLong();
            if (!ExistsQQ(botKey, recUserId))
                return "此号码未登记，请确认号码正确";

            if (BlackList.IsSystemBlack(recUserId))
                return "此号码已被列入官方黑名单";

            if (await UserInfo.AppendAsync(botUin, groupId, userId, name, GroupInfo.GetRobotOwner(groupId)) == -1)
                return RetryMsg;

            long creditAdd = 5000;
            long creditValue = await UserInfo.GetCreditAsync(userId);
            long creditValue2 = await UserInfo.GetCreditAsync(recUserId);

            using var trans = await BeginTransactionAsync();
            try
            {
                // 1. 更新推荐人
                var (sqlRec, parasRec) = SqlUpdateWhere($"RecUserId = {recUserId}, BindCredit = BindCredit + 5000 ", $"Id = {clientOid}");
                await ExecAsync(sqlRec, trans, parasRec);

                // 2. 自己加分
                var addResSelf = await UserInfo.AddCreditAsync(botUin, groupId, groupName, userId, name, creditAdd, "推荐关注", trans);
                if (addResSelf.Result == -1) throw new Exception("自己加分失败");

                // 3. 推荐人加分
                var addResRec = await UserInfo.AddCreditAsync(botUin, groupId, groupName, recUserId, "", creditAdd, $"推荐关注:{userId}", trans);
                if (addResRec.Result == -1) throw new Exception("推荐人加分失败");

                await trans.CommitAsync();

                // 同步缓存
                UserInfo.SyncCacheField(userId, groupId, "Credit", addResSelf.CreditValue);
                UserInfo.SyncCacheField(recUserId, groupId, "Credit", addResRec.CreditValue);

                return $"推荐人登记为：\n{recUserId} +{creditAdd}分，累计：{addResRec.CreditValue}\n您的积分：{creditAdd}，累计：{addResSelf.CreditValue}";
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                Console.WriteLine($"[GetRecRes Error] {ex.Message}");
                return RetryMsg;
            }
        }

        public static string GetBindToken(BotMessage bm, string tokenType, string bindToken)
        {
            return GetBindTokenAsync(bm, tokenType, bindToken).GetAwaiter().GetResult();
        }

        public static async Task<string> GetBindTokenAsync(BotMessage bm, string tokenType, string bindToken)
        {
            string res = "";
            long creditAdd = 5000;
            if (bindToken.Trim() == "") return "";
            var botUin = bm.SelfId;
            var groupId = bm.GroupId;
            var groupName = bm.GroupName;
            var UserId = bm.UserId;
            var name = bm.Name;

            if (tokenType == "MP")
            {
                if (bm.IsPublic)
                    return "请用QQ发消息领取，体验群：6433316";

                var resObj = QueryScalar<object>($"select {SqlTop(1)} 1 from {FullName} where BotKey = (select {SqlTop(1)} BotKey from {FullName} where BindToken = {bindToken.Quotes()}{SqlLimit(1)}) and UserId = {UserId}{SqlLimit(1)}");
                if (resObj != null && resObj != DBNull.Value)
                    return "您已领过积分，不能再次领取";

                res = QueryScalar<string>($"select UserId from {FullName} where BindToken = {bindToken.Quotes()}") ?? "";
                if (res.IsNull())
                    return "此TOKEN无效，请确认后再试";

                long bindUserId = res.AsLong();
                if (IsBind(bindUserId))
                    return "此TOKEN已失效，请重新获取";

                bm.AddClient();

                //更新绑定信息，加入事务运行 更新积分记录 发送消息记录
                long creditValue = await UserInfo.GetCreditAsync(groupId, UserId);

                using var trans = await BeginTransactionAsync();
                try
                {
                    // 1. 更新绑定信息
                    var (sqlBind, parasBind) = SqlUpdateWhere($"UserId = {UserId}, IsBind = 1, BindDate = {SqlDateTime}, BindCredit = {creditAdd}", $"BindToken = {bindToken.Quotes()}");
                    await ExecAsync(sqlBind, trans, parasBind);

                    // 2. 更新积分记录关联
                    await ExecAsync($"UPDATE {CreditLog.FullName} SET UserId = {UserId} WHERE UserId = {bindUserId}", trans);

                    // 3. 更新发送消息记录关联
                    await ExecAsync($"UPDATE {GroupSendMessage.FullName} SET UserId = {UserId} WHERE UserId = {bindUserId}", trans);

                    // 4. 处理 Token 表
                    if (Token.Exists(UserId))
                    {
                        await ExecAsync($"DELETE FROM {Token.FullName} WHERE UserId = {bindUserId}", trans);
                    }
                    else
                    {
                        await ExecAsync($"UPDATE {Token.FullName} SET UserId = {UserId} WHERE UserId = {bindUserId}", trans);
                    }

                    // 5. 增加积分
                    var addRes = await UserInfo.AddCreditAsync(botUin, groupId, groupName, UserId, name, creditAdd, "关注公众号领积分", trans);
                    if (addRes.Result == -1) throw new Exception("增加积分失败");

                    await trans.CommitAsync();

                    // 同步缓存
                    UserInfo.SyncCacheField(UserId, groupId, "Credit", addRes.CreditValue);

                    return $"得分：{creditAdd}，累计：{addRes.CreditValue}";
                }
                catch (Exception ex)
                {
                    await trans.RollbackAsync();
                    Console.WriteLine($"[GetBindToken Error] {ex.Message}");
                    return RetryMsg;
                }
            }
            else if (tokenType == "WX")
            {
                return "功能尚未实现，请稍后";
            }
            else if (tokenType == "QQ")
            {
                //QQ私聊得到TOKEN，发到微信群或微信公众号领取分（微信机器人的在PYTHON代码实现，公众号的在……）
                return "功能尚未实现，请稍后";
            }
            return res;
        }

        public static int Append(string botKey, string clientKey)
        {
            int i = Insert([
                new Cov("BotKey", botKey),
                new Cov("UserKey", clientKey),
                new Cov("BindToken", GetBindToken(botKey, clientKey)),
            ]);
            if (i == -1)
                return i;
            else
            {
                long clientOid = GetId(botKey, clientKey);
                return Update($"UserId = {startUserId} + Id", clientOid);
            }
        }

    }

}
