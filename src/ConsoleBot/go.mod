module consolebot

go 1.25.0

require (
	BotMatrix/common v0.0.0
	botworker v0.0.0
)

replace BotMatrix/common => ../Common
replace botworker => ../BotWorker
