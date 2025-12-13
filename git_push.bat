@echo off
set /p msg="Enter commit message: "
if "%msg%"=="" set msg="Update"

git add .
git commit -m "%msg%"
git push

echo.
echo Done!
pause