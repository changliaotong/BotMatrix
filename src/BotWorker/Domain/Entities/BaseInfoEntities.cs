using Dapper.Contrib.Extensions;

namespace BotWorker.Domain.Entities
{
    [Table("chengyu")]
    public partial class Chengyu
    {
        [Key]
        public long Oid { get; set; }
        public string Name { get; set; } = string.Empty;
        public string Pingyin { get; set; } = string.Empty;
        public string Pinyin { get; set; } = string.Empty;
        public string Spinyin { get; set; } = string.Empty;
        public string Diangu { get; set; } = string.Empty;
        public string Chuchu { get; set; } = string.Empty;
        public string Lizi { get; set; } = string.Empty;
    }

    [Table("ciba")]
    public partial class Cidian
    {
        [ExplicitKey]
        public string Keyword { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
    }

    [Table("city")]
    public partial class City
    {
        [Key]
        public int Id { get; set; }
        public string CityName { get; set; } = string.Empty;
        public string Province { get; set; } = string.Empty;
        public string AreaCode { get; set; } = string.Empty;
        public string ZipCode { get; set; } = string.Empty;
    }
}
