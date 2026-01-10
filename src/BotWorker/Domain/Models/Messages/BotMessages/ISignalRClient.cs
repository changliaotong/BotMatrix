namespace BotWorker.Domain.Models.Messages.BotMessages
{
    public interface ISignalRClient
    {
        Task MuteAsync(
            string userId,
            string botMessageJson,
            long groupId,
            long targetId,
            int seconds);

        Task KickOutAsync(
            string userId,
            string botMessageJson,
            long groupId,
            long targetId);

        Task ChangeNameAsync(
            string userId,
            string botMessageJson,
            long groupId,
            long targetUin,
            string newName,
            string prefixBoy,
            string prefixGirl,
            string prefixAdmin);

        Task ChangeNameAllAsync(
            string userId,
            string botMessageJson,
            string prefixBoy,
            string prefixGirl,
            string prefixAdmin);

        Task RecallAsync(
            string userId,
            string botMessageJson,
            string msgGuid,
            long groupId,
            string msgId);

        Task RecallForward(
            string userId,
            string botMessageJson,
            string group,
            string msgId,
            string forwardId);

        Task SetTitleAsync(
            string userId,
            string botMessageJson,
            long groupId,
            long targetUin,
            string title);

        Task LeaveAsync(
            string userId,
            string botMessageJson,
            long groupId);
    }    
}