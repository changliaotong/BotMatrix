package plugins

import (
	"botworker/internal/config"
	"botworker/internal/plugin"
	"database/sql"
)

// Stub plugin structure
type stubPlugin struct {
	name string
}

func (p *stubPlugin) Name() string        { return p.name }
func (p *stubPlugin) Description() string { return p.name + " stub plugin" }
func (p *stubPlugin) Version() string     { return "1.0.0" }
func (p *stubPlugin) Init(robot plugin.Robot) {
	// Do nothing
}

func NewWeatherPlugin(cfg *config.WeatherConfig) plugin.Plugin {
	return &stubPlugin{name: "Weather"}
}

func NewPointsPlugin(db *sql.DB) plugin.Plugin {
	return &stubPlugin{name: "Points"}
}

func NewSignInPlugin(pointsPlugin plugin.Plugin) plugin.Plugin {
	return &stubPlugin{name: "SignIn"}
}

func NewAuctionPlugin(db *sql.DB, pointsPlugin plugin.Plugin) plugin.Plugin {
	return &stubPlugin{name: "Auction"}
}

func NewMedalPlugin() plugin.Plugin {
	return &stubPlugin{name: "Medal"}
}

func NewGamesPlugin() plugin.Plugin {
	return &stubPlugin{name: "Games"}
}

func NewLotteryPlugin() plugin.Plugin {
	return &stubPlugin{name: "Lottery"}
}

func NewMenuPlugin() plugin.Plugin {
	return &stubPlugin{name: "Menu"}
}

func NewTranslatePlugin(cfg *config.TranslateConfig) plugin.Plugin {
	return &stubPlugin{name: "Translate"}
}

func NewMusicPlugin() plugin.Plugin {
	return &stubPlugin{name: "Music"}
}

func NewPetPlugin(db *sql.DB, pointsPlugin plugin.Plugin) plugin.Plugin {
	return &stubPlugin{name: "Pet"}
}

func NewTimePlugin() plugin.Plugin {
	return &stubPlugin{name: "Time"}
}

func NewQAPlugin() plugin.Plugin {
	return &stubPlugin{name: "QA"}
}

func NewGiftPlugin(db *sql.DB) plugin.Plugin {
	return &stubPlugin{name: "Gift"}
}

func NewMarriagePlugin() plugin.Plugin {
	return &stubPlugin{name: "Marriage"}
}

func NewBabyPlugin() plugin.Plugin {
	return &stubPlugin{name: "Baby"}
}

func NewBadgePlugin() plugin.Plugin {
	return &stubPlugin{name: "Badge"}
}

func NewSmallGamesPlugin() plugin.Plugin {
	return &stubPlugin{name: "SmallGames"}
}

func NewKnowledgeBasePlugin(db *sql.DB, groupID string) plugin.Plugin {
	return &stubPlugin{name: "KnowledgeBase"}
}

func NewDialogDemoPlugin() plugin.Plugin {
	return &stubPlugin{name: "DialogDemo"}
}

func NewTestServerPlugin() plugin.Plugin {
	return &stubPlugin{name: "TestServer"}
}

func NewRobberyPlugin(db *sql.DB) plugin.Plugin {
	return &stubPlugin{name: "Robbery"}
}

func NewFishingPlugin(db *sql.DB) plugin.Plugin {
	return &stubPlugin{name: "Fishing"}
}

func NewSystemInfoPlugin() plugin.Plugin {
	return &stubPlugin{name: "SystemInfo"}
}

func NewCultivationPlugin(db *sql.DB) plugin.Plugin {
	return &stubPlugin{name: "Cultivation"}
}

func NewFarmPlugin(db *sql.DB) plugin.Plugin {
	return &stubPlugin{name: "Farm"}
}

func NewTarotPlugin() plugin.Plugin {
	return &stubPlugin{name: "Tarot"}
}

func NewWordGuessPlugin() plugin.Plugin {
	return &stubPlugin{name: "WordGuess"}
}

func NewIdiomGuessPlugin() plugin.Plugin {
	return &stubPlugin{name: "IdiomGuess"}
}

func NewPluginManagerPlugin(db *sql.DB) plugin.Plugin {
	return &stubPlugin{name: "PluginManager"}
}
