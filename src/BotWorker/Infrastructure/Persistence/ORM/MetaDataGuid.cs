using Newtonsoft.Json;
using BotWorker.Infrastructure.Extensions;

namespace BotWorker.Infrastructure.Persistence.ORM
{
    public abstract class MetaDataGuid<TDerived> : MetaData<TDerived> where TDerived : MetaDataGuid<TDerived>, new()
    {
        public long Id { get; set; }
        public Guid Guid { get; set; } = default;

        protected virtual string GuidFieldName => "Guid";
        protected virtual string IdFieldName => "Id";

        public static string GuidField { get; private set; }
        public static string IdField { get; private set; }

        static MetaDataGuid()
        {
            var instance = new TDerived();            
            GuidField = instance.GuidFieldName;
            IdField = instance.IdFieldName;
        }

        public static async Task<TDerived?> LoadAsync(Guid guid)
        {
            if (Key.Equals(GuidField, StringComparison.OrdinalIgnoreCase))
                return await GetSingleAsync(guid);
            else
                return await GetSingleAsync(GetId(guid));
        }

        public static async Task<TDerived?> LoadAsync(long Id)
        {
            if (Key.Equals(IdField, StringComparison.OrdinalIgnoreCase))            
                return await GetSingleAsync(Id);            
            else
                return await GetSingleAsync(GetGuid(Id));
        }

        public static long GetId(Guid guid)
        {  
            if (Key.Equals(GuidField, StringComparison.OrdinalIgnoreCase))            
                return GetLong(IdField, guid);            
            else            
                return GetWhere<long>(IdField, $"{GuidField} = {guid.AsString().Quotes()}");            
        }

        public static long GetId(string guid)
        {
            return GetId(Guid.Parse(guid));
        }

        public static Guid GetGuid(long id)
        {
            if (Key.Equals(IdField, StringComparison.OrdinalIgnoreCase))            
                return Get<Guid>(GuidField, id);            
            else            
                return GetWhere<Guid>(GuidField, $"{IdField} = {id}");            
        }

        public static Dictionary<string, object?>? GetDict(Guid guid, params string[] fields)
        {
            return GetDicts($"{GuidField} = @guid", SqlParams(("@guid", guid)) , fields).FirstOrDefault();
        }

        public static Dictionary<string, object?>? GetDict(long id, params string[] fields)
        {
            return GetDicts($"{IdField} = @id", SqlParams(("@id", id)), fields).FirstOrDefault();
        }
    }
}

