@echo off
echo [Git] Adding all files...
git add .
echo [Git] Committing...
set /p msg="Enter commit message (default: Update): "
if "%msg%"=="" set msg="Update"
git commit -m "%msg%"
echo [Git] Pushing to main...
git push origin main
echo [Git] Done!
pause