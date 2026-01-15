using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Persistence.Repositories;

namespace BotWorker.Domain.Models.BotMessages
{
    public partial class BotMessage
    {
        // 临时使用静态属性，后续应通过依赖注入解决
        public static IUserRepository UserRepository { get; } = new UserRepository();
        public static IGroupRepository GroupRepository { get; } = new GroupRepository();
        public static IGroupMemberRepository GroupMemberRepository { get; } = new GroupMemberRepository();
        public static ISignInRepository SignInRepository { get; } = new SignInRepository();
        public static IBotRepository BotRepository { get; } = new BotRepository();
        public static ITokensLogRepository TokenLogRepository { get; } = new TokenLogRepository();
        public static ICreditLogRepository CreditLogRepository { get; } = new CreditLogRepository();
    }
}
