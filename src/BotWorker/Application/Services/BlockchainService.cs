using System.Threading.Tasks;

namespace BotWorker.Application.Services
{
    public interface IBlockchainService
    {
        Task<string> GetBalanceAsync(string address, string chain);
        Task<string> TransferAsync(string from, string to, decimal amount, string chain);
    }

    public class BlockchainService : IBlockchainService
    {
        public async Task<string> GetBalanceAsync(string address, string chain)
        {
            return await Task.FromResult("0.00");
        }

        public async Task<string> TransferAsync(string from, string to, decimal amount, string chain)
        {
            return await Task.FromResult("Transaction Hash: 0x...");
        }
    }
}


