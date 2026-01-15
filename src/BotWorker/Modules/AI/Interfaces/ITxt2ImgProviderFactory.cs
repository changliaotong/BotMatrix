namespace BotWorker.Modules.AI.Interfaces
{
    public interface ITxt2ImgProviderFactory
    {
        ITxt2ImgProvider? CreateProvider(string? providerName = null);
    }
}
