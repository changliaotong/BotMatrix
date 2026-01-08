using Microsoft.AspNetCore.Components;

namespace BotWorker.Core.Services
{
    public class ApiHttpClient
    {
        public HttpClient Client { get; }

        public ApiHttpClient(IHttpClientFactory factory, NavigationManager nav)
        {
            Client = factory.CreateClient("DefaultClient");
            Client.BaseAddress = new Uri(nav.BaseUri);
            Client.DefaultRequestHeaders.Add("X-Api-Key", apiKey);
        }
    }

}


