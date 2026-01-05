module github.com/BotMatrix/src/tools/bm-cli

go 1.25.0

replace github.com/BotMatrix/src/Common => ../../Common
replace BotMatrix/common => ../../Common

require (
	github.com/fsnotify/fsnotify v1.7.0
	BotMatrix/common v0.0.0
)
