using Microsoft.AspNetCore.SignalR;

namespace sz84.Infrastructure.WebRTC
{
    public class WebRtcHub : Hub
    {
        public async Task SendOffer(string targetUserId, string sdp)
        {
            await Clients.User(targetUserId).SendAsync("ReceiveOffer", Context.UserIdentifier, sdp);
        }

        public async Task SendAnswer(string targetUserId, string sdp)
        {
            await Clients.User(targetUserId).SendAsync("ReceiveAnswer", Context.UserIdentifier, sdp);
        }

        public async Task SendIceCandidate(string targetUserId, string candidate)
        {
            await Clients.User(targetUserId).SendAsync("ReceiveIceCandidate", Context.UserIdentifier, candidate);
        }

        public override Task OnConnectedAsync()
        {
            Console.WriteLine($"User connected: {Context.UserIdentifier}");
            return base.OnConnectedAsync();
        }
    }

}
