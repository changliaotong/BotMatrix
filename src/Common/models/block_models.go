package models

import (
	"time"
)

// Block 区块链游戏记录
type Block struct {
	Id           int64     `gorm:"column:Id;primaryKey;autoIncrement" json:"id"`
	PrevId       int64     `gorm:"column:PrevId" json:"prev_id"`
	PrevHash     string    `gorm:"column:PrevHash;size:64" json:"prev_hash"`
	PrevRes      string    `gorm:"column:PrevRes;size:255" json:"prev_res"`
	BotUin       int64     `gorm:"column:BotUin" json:"bot_uin"`
	GroupId      int64     `gorm:"column:GroupId" json:"group_id"`
	GroupName    string    `gorm:"column:GroupName;size:255" json:"group_name"`
	UserId       int64     `gorm:"column:UserId" json:"user_id"`
	UserName     string    `gorm:"column:UserName;size:255" json:"user_name"`
	BlockInfo    string    `gorm:"column:BlockInfo;type:text" json:"block_info"`
	BlockSecret  string    `gorm:"column:BlockSecret;type:text" json:"block_secret"`
	BlockNum     int       `gorm:"column:BlockNum" json:"block_num"`
	BlockRes     string    `gorm:"column:BlockRes;size:255" json:"block_res"`
	BlockRand    string    `gorm:"column:BlockRand;size:64" json:"block_rand"`
	BlockHash    string    `gorm:"column:BlockHash;size:64;index" json:"block_hash"`
	IsOpen       bool      `gorm:"column:IsOpen" json:"is_open"`
	OpenDate     time.Time `gorm:"column:OpenDate" json:"open_date"`
	OpenBotUin   int64     `gorm:"column:OpenBotUin" json:"open_bot_uin"`
	OpenUserId   int64     `gorm:"column:OpenUserId" json:"open_user_id"`
	OpenUserName string    `gorm:"column:OpenUserName;size:255" json:"open_user_name"`
	InsertDate   time.Time `gorm:"column:InsertDate;autoCreateTime" json:"insert_date"`
}

func (Block) TableName() string {
	return "Block"
}

// BlockRandom 随机数表 (预生成的骰子点数 1-6 * 3)
type BlockRandom struct {
	Id       int64 `gorm:"column:Id;primaryKey;autoIncrement" json:"id"`
	BlockNum int   `gorm:"column:BlockNum" json:"block_num"`
}

func (BlockRandom) TableName() string {
	return "BlockRandom"
}

// BlockType 游戏类型 (押大、押小、押单、押双等)
type BlockType struct {
	Id        int64  `gorm:"column:Id;primaryKey;autoIncrement" json:"id"`
	TypeName  string `gorm:"column:TypeName;size:50" json:"type_name"`
	BlockOdds int    `gorm:"column:BlockOdds" json:"block_odds"`
}

func (BlockType) TableName() string {
	return "Blocktype" // Note: Case sensitive from C# "Blocktype"
}

// BlockWin 输赢规则表
type BlockWin struct {
	Id       int64 `gorm:"column:Id;primaryKey;autoIncrement" json:"id"`
	TypeId   int64 `gorm:"column:TypeId" json:"type_id"`
	BlockNum int   `gorm:"column:BlockNum" json:"block_num"`
	IsWin    bool  `gorm:"column:IsWin" json:"is_win"`
}

func (BlockWin) TableName() string {
	return "BlockWin"
}
