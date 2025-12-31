# Baby and Marriage System

> [🌐 English](BABY_AND_MARRIAGE_SYSTEM.md) | [简体中文](../zh-CN/BABY_AND_MARRIAGE_SYSTEM.md)
> [⬅️ Back to Docs](README.md) | [🏠 Back to Home](../../README.md)

## 1. System Overview

BotMatrix implements two core interaction systems: the **Baby System** and the **Marriage System**. These systems are designed to enhance the interactivity between the bot and users, providing a rich virtual social experience.

### 1.1 Baby System
The Baby System allows users to own and nurture virtual babies, including baby arrival, learning, working, and interaction. A growth value system tracks the baby's development.

### 1.2 Marriage System
The Marriage System allows users to find partners in the virtual world, with features for proposing, marrying, and divorcing. It also includes interactive elements like wedding sweets, red packets, and "sweet hearts."

## 2. Data Models

### 2.1 Baby System Data Models

#### Baby - Basic Information
```go
type Baby struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    UserID      string    `gorm:"size:20;index" json:"user_id"`
    Name        string    `gorm:"size:50" json:"name"`
    Birthday    time.Time `json:"birthday"`
    GrowthValue int       `json:"growth_value"`
    DaysOld     int       `json:"days_old"`
    Level       int       `json:"level"`
    Status      string    `gorm:"size:20;default:active" json:"status"` // active, abandoned
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

#### BabyEvent - Event Records
```go
type BabyEvent struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    BabyID    uint      `json:"baby_id"`
    EventType string    `gorm:"size:50" json:"event_type"` // birthday, learn, work, interact
    Content   string    `gorm:"size:255" json:"content"`
    CreatedAt time.Time `json:"created_at"`
}
```

#### BabyConfig - System Configuration
```go
type BabyConfig struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    IsEnabled   bool      `gorm:"default:true" json:"is_enabled"`
    GrowthRate  int       `gorm:"default:1000" json:"growth_rate"` // e.g., 1 day per 1000 growth
    UpdateAt    time.Time `json:"update_at"`
}
```

### 2.2 Marriage System Data Models

#### UserMarriage - Marriage Information
```go
type UserMarriage struct {
    ID              uint      `gorm:"primaryKey" json:"id"`
    UserID          string    `gorm:"size:20;index" json:"user_id"`
    SpouseID        string    `gorm:"size:20;index" json:"spouse_id"`
    MarriageDate    time.Time `json:"marriage_date"`
    DivorceDate     time.Time `json:"divorce_date"`
    Status          string    `gorm:"size:20;default:single" json:"status"` // single, married, divorced
    SweetsCount     int       `gorm:"default:0" json:"sweets_count"`
    RedPacketsCount int       `gorm:"default:0" json:"red_packets_count"`
    SweetHearts     int       `gorm:"default:0" json:"sweet_hearts"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}
```

#### MarriageProposal - Proposal Records
```go
type MarriageProposal struct {
    ID           uint      `gorm:"primaryKey" json:"id"`
    ProposerID   string    `gorm:"size:20;index" json:"proposer_id"`
    RecipientID  string    `gorm:"size:20;index" json:"recipient_id"`
    Status       string    `gorm:"size:20;default:pending" json:"status"` // pending, accepted, rejected
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

#### WeddingItem - Wedding Items
```go
type WeddingItem struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    ItemType    string    `gorm:"size:20" json:"item_type"` // dress, ring
    Name        string    `gorm:"size:50" json:"name"`
    Price       int       `gorm:"default:0" json:"price"`
    Description string    `gorm:"size:255" json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```
