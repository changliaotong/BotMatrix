using System;

namespace BotWorker.Core.Database.Mapping
{
    [AttributeUsage(AttributeTargets.Property, AllowMultiple = false)]
    public class DisplayInListAttribute(bool show) : Attribute
    {
        public bool Show { get; } = show;
    }

    [AttributeUsage(AttributeTargets.Property, AllowMultiple = false)]
    public class DisplayInEditAttribute(bool show) : Attribute
    {
        public bool Show { get; } = show;
    }
}