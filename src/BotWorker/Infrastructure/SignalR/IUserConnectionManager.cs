
namespace sz84.Infrastructure.SignalR
{
    public interface IUserConnectionManager
    {
        void AddConnection(string userId, string connectionId, string role = "user");
        void RemoveConnection(string connectionId);
        void UpdateActivity(string connectionId);
        bool IsOnline(string userId);
        List<string> GetConnections(string userId);
        string? GetSingleConnection(string userId);
        List<string> GetConnectionsByRole(string role);
        List<string> GetOnlineUserIds();
        int OnlineUserCount { get; }
        int GetRoleCount(string role);
        void RemoveInactiveConnections(TimeSpan timeout);
        bool CanSend(string connectionId);
    }

}
