namespace sz84.Core.Services
{
    public class PythonApiClient
    {
        public HttpClient Client { get; }

        public PythonApiClient(IHttpClientFactory factory)
        {
            Client = factory.CreateClient("PythonApiClient");
        }
    }
}
