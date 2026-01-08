using System;
using System.Threading.Tasks;
using BotWorker.Domain.Models.Messages.BotMessages;
using BotWorker.Common.Extensions;

namespace BotWorker.Application.Messaging.Handlers
{
    public class GameCommandHandler
    {
        public async Task<CommandResult> HandleAsync(BotMessage botMsg)
        {
            if (botMsg.InGame())
            {
                botMsg.CmdName = "成语接龙";
                botMsg.CmdPara = botMsg.CurrentMessage;
                var res = await botMsg.GetJielongRes();
                if (!string.IsNullOrEmpty(res))
                    return CommandResult.Intercepted(res);
            }

            return CommandResult.Continue();
        }
    }
}


