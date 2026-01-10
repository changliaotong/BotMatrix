using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Utils;

namespace BotWorker.Application.Messaging.Handlers
{
    public class AdminCommandHandler
    {
        private readonly Services.IGroupService _groupService;

        public AdminCommandHandler(Services.IGroupService groupService)
        {
            _groupService = groupService;
        }

        public Task<CommandResult> HandleAsync(BotMessage botMsg)
        {
            // 已由 AdminService 插件接管
            return Task.FromResult(CommandResult.Continue());
        }
    }
}
