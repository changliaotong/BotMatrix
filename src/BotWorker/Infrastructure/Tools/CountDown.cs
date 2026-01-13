namespace BotWorker.Infrastructure.Tools
{
    public class CountDown : MetaData<CountDown>
    {
        public override string TableName => throw new NotImplementedException();

        public override string KeyField => throw new NotImplementedException();

        // 倒计时
        public static async Task<string> GetCountDownAsync()
        {
            string sql = "SELECT DATEDIFF(DAY, GETDATE(), '2025-10-01'), DATEDIFF(DAY, GETDATE(), '2025-10-06'), DATEDIFF(DAY, GETDATE(), '2026-01-01'), DATEDIFF(DAY, GETDATE(), '2026-02-17'), DATEDIFF(DAY, GETDATE(), '2026-06-07')";
            return await QueryResAsync(sql, "🕒 2025倒计时：\n🇨🇳 国庆节{0}天✨(25/10/01)\n🌕 中秋节{1}天🥮（25/10/06）\n\n🕒 2026倒计时：\n✨ 元旦{2}天🎉（26/01/01）\n🏮 春节{3}天🧨（26/02/17）\n📚 高考{4}天✏️（26/06/07）");
        }

        public static string GetCountDown() => GetCountDownAsync().GetAwaiter().GetResult();
    }
}
