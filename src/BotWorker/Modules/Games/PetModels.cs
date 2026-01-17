using System;
using System.Collections.Generic;
using Dapper.Contrib.Extensions;

namespace BotWorker.Modules.Games
{
    #region Attributes & Config

    [AttributeUsage(AttributeTargets.Method)]
    public class PetCommandAttribute : Attribute
    {
        public string[] Aliases { get; }
        public string Description { get; }
        public int Order { get; }
        public PetCommandAttribute(string[] aliases, string description, int order = 0)
        { 
            Aliases = aliases;
            Description = description;
            Order = order;
        }
    }

    public class PetConfig
    {
        public double ExpMultiplier { get; set; } = 1.0;
        public double HungerRate { get; set; } = 2.0;
        public double EnergyRecoveryRate { get; set; } = 5.0;
        public int MaxPetCount { get; set; } = 1;
        public string DefaultPetName { get; set; } = "小萌新";
        public double IntimacyGainRate { get; set; } = 1.0;
    }

    #endregion

    #region Domain Model

    [Table("UserPets")]
    public class Pet
    {
        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string UserId { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public PetType Type { get; set; } = PetType.Cat;
        public PetPersonality Personality { get; set; } = PetPersonality.Ordinary;
        public PetState CurrentState { get; set; } = PetState.Idle;
        public DateTime StateEndTime { get; set; } = DateTime.MinValue;

        public DateTime AdoptTime { get; set; }
        public DateTime LastUpdateTime { get; set; }

        public double Health { get; set; } = 100;
        public double Hunger { get; set; } = 0;
        public double Happiness { get; set; } = 100;
        public double Energy { get; set; } = 100;
        public double Intimacy { get; set; } = 0; // 亲密度
        public int Gold { get; set; } = 100; // 金币
        public int Level { get; set; } = 1;
        public double Experience { get; set; } = 0;

        [Write(false)]
        [Computed]
        public double ExperienceToNextLevel => 100 * Math.Pow(Level, 1.2);

        [Write(false)]
        [Computed]
        public int Age => (DateTime.Now - AdoptTime).Days;

        [Write(false)]
        [Computed]
        public string PersonalityName => Personality switch
        {
            PetPersonality.Ordinary => "平凡",
            PetPersonality.Energetic => "精力充沛",
            PetPersonality.Lazy => "懒散",
            PetPersonality.Aggressive => "好斗",
            PetPersonality.Gentle => "温柔",
            _ => "未知"
        };

        [Write(false)]
        [Computed]
        public List<string> Events { get; } = new();
    }

    [Table("pet_inventory")]
    public class PetItem
    {
        [ExplicitKey]
        public Guid Id { get; set; } = Guid.NewGuid();
        public string UserId { get; set; } = string.Empty;
        public string ItemId { get; set; } = string.Empty;
        public int Count { get; set; }
    }

    #endregion

    #region Enums

    public enum PetType { Cat, Dog, Bird, Rabbit, Dragon }
    public enum PetPersonality { Ordinary, Energetic, Lazy, Aggressive, Gentle }
    public enum PetState { Idle, Sleeping, Working, Playing, Training, Exploring }

    #endregion
}