using System;

namespace BotWorker.Infrastructure.Utils.Schema.Attributes;

[AttributeUsage(AttributeTargets.Property)]
public class ColumnAttribute : Attribute
{
    public string Name { get; }

    public ColumnAttribute(string name)
    {
        Name = name;
    }
}