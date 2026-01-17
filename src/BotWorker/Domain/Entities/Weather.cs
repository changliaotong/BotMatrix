using System;

namespace BotWorker.Domain.Entities
{
    public class Weather
    {
        public int Id { get; set; }
        public string CityName { get; set; } = string.Empty;
        public string WeatherInfo { get; set; } = string.Empty;
        public DateTime InsertDate { get; set; }
    }
}
