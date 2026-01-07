namespace BotWorker.Bots.Models.Buses.Data
{
    public class StopPlace
    {
        public int stop_id { get; set; }
        public string? stop_name { get; set; }

        public int place_id { get; set; }

        public string? place_name { get; set; }

        public int search_times { get; set; }
    }
}
