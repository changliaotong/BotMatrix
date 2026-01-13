@echo off
:: Resolve UU conflicts (Keep Ours)
git checkout --ours go.work.sum
git add go.work.sum
git checkout --ours src/Common/ai/ai_service.go
git add src/Common/ai/ai_service.go
git checkout --ours src/Common/ai/employee/init_templates.go
git add src/Common/ai/employee/init_templates.go
git checkout --ours src/Common/ai/handlers_ai_mgmt.go
git add src/Common/ai/handlers_ai_mgmt.go
git checkout --ours src/Common/config/loader.go
git add src/Common/config/loader.go
git checkout --ours src/Common/database/db.go
git add src/Common/database/db.go
git checkout --ours src/Common/models/gorm_models.go
git add src/Common/models/gorm_models.go
git checkout --ours src/WebUI/src/components/layout/PortalHeader.vue
git add src/WebUI/src/components/layout/PortalHeader.vue
git checkout --ours src/WebUI/src/router/index.ts
git add src/WebUI/src/router/index.ts
git checkout --ours src/WebUI/node_modules/.vite/deps/_metadata.json
git add src/WebUI/node_modules/.vite/deps/_metadata.json

:: Resolve BotWorker UU conflicts (Keep Theirs)
git checkout --theirs src/BotWorker/Modules/Games/AdminService.cs
git add src/BotWorker/Modules/Games/AdminService.cs
git checkout --theirs src/BotWorker/Modules/Games/SimpleGame.cs
git add src/BotWorker/Modules/Games/SimpleGame.cs
git checkout --theirs src/BotWorker/Program.cs
git add src/BotWorker/Program.cs
git checkout --theirs src/BotWorker/appsettings.json
git add src/BotWorker/appsettings.json

:: Resolve DU conflicts (Keep Theirs for BotWorker, Keep Ours for others)
for /f "tokens=2" %%i in ('git status --short ^| findstr "^DU"') do (
    echo %%i | findstr "src/BotWorker" > nul
    if errorlevel 1 (
        git rm %%i
    ) else (
        git checkout --theirs %%i
        git add %%i
    )
)

:: Final check
git status --short
