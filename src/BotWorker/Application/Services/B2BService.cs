using System.Threading.Tasks;

namespace BotWorker.Application.Services
{
    public interface IB2BService
    {
        Task<string> CreateOrderAsync(string merchantId, decimal amount);
        Task<bool> QueryOrderAsync(string orderId);
    }

    public class B2BService : IB2BService
    {
        public async Task<string> CreateOrderAsync(string merchantId, decimal amount)
        {
            return await Task.FromResult($"ORDER-{Guid.NewGuid()}");
        }

        public async Task<bool> QueryOrderAsync(string orderId)
        {
            return await Task.FromResult(true);
        }
    }
}


