namespace BotWorker.Bots.Models.Buses.Data
{
    public class StopBus
    {
        public int bus_id { get; set; }

        public string? bus_name { get; set; }

        public string? stop_no { get; set; }

        public string? stop_no2 { get; set; }

        public int next_stop { get; set; }

        public int next_stop2 { get; set; }


        public string? bus_time { get; set; }

        public string? bus_price { get; set; }
    }
}
