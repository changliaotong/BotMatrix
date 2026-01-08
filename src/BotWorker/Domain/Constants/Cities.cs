namespace BotWorker.Domain.Constants
{
    /// <summary>
    /// �й���������
    /// </summary>
    public class Cities : MetaData<Cities>
    {
        public override string DataBase => "baseinfo";
        public override string TableName => "cities";
        public override string KeyField => "city_name";
    }
}


