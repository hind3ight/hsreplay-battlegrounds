@echo off
chcp 65001 >nul
echo ========================================
echo Battlegrounds Advisor - Go Build Script
echo ========================================
echo.

:: 切换到项目根目录
cd /d "%~dp0.."
echo [Debug] Working dir: %CD%
echo [Debug] Looking for go.mod...
if not exist go.mod (
    echo [Error] go.mod not found in %CD%
    pause
    exit /b 1
)
echo [OK] go.mod found
echo.

where go >nul 2>&1
if errorlevel 1 (
    echo [Error] Go not found. Install from https://go.dev
    pause
    exit /b 1
)

for /f "tokens=*" %%i in ('go version') do set GOVER=%%i
echo [Info] %GOVER%
echo.

echo [Step 1 of 2] Tidy and download Go modules...
go mod tidy
if errorlevel 1 (
    echo [Error] go mod tidy failed
    pause
    exit /b 1
)
echo [Done]
echo.

echo [Step 2 of 2] Building executables...
mkdir bin 2>nul

echo   Building fetch.exe...
go build -ldflags="-H=windowsgui" -o bin/fetch.exe cmd/fetch/main.go
if errorlevel 1 (
    echo [Error] fetch.exe build failed
    pause
    exit /b 1
)
echo   [OK]

echo   Building analyze.exe...
go build -ldflags="-H=windowsgui" -o bin/analyze.exe cmd/analyze/main.go
if errorlevel 1 (
    echo [Error] analyze.exe build failed
    pause
    exit /b 1
)
echo   [OK]

echo   Building interactive.exe...
go build -ldflags="-H=windowsgui" -o bin/interactive.exe cmd/interactive/main.go
if errorlevel 1 (
    echo [Error] interactive.exe build failed
    pause
    exit /b 1
)
echo   [OK]

echo   Building analyze-cross.exe...
go build -ldflags="-H=windowsgui" -o bin/analyze-cross.exe cmd/analyze-cross/main.go
if errorlevel 1 (
    echo [Error] analyze-cross.exe build failed
    pause
    exit /b 1
)
echo   [OK]

echo.
echo ========================================
echo All 4 executables built successfully.
echo Output: bin\
echo ========================================
dir bin\*.exe
pause
