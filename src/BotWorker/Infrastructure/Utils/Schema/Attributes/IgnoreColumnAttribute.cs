using System;

namespace BotWorker.Infrastructure.Utils.Schema.Attributes;

[AttributeUsage(AttributeTargets.Property)]
public class IgnoreColumnAttribute : Attribute
{
}