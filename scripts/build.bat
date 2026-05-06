@echo off
chcp 65001 >nul
echo ========================================
echo Battlegrounds Advisor - Go Build Script
echo ========================================
echo.

:: 切换到项目根目录
cd /d "%~dp0.."

where go >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [错误] 未找到 Go，请从 https://go.dev 安装
    pause
    exit /b 1
)

for /f "tokens=*" %%i in ('go version') do set GOVER=%%i
echo [信息] %GOVER%
echo.

:: 检查依赖
echo [Step 1 of 2] Installing Go dependencies...
go mod download
echo [Done]
echo.

:: 编译 Windows amd64 二进制
echo [Step 2 of 2] Building Windows executables...
mkdir bin 2>nul

echo   Building fetch.exe...
go build -ldflags="-H=windowsgui" -o bin/fetch.exe cmd/fetch/main.go

echo   Building analyze.exe...
go build -ldflags="-H=windowsgui" -o bin/analyze.exe cmd/analyze/main.go

echo   Building interactive.exe...
go build -ldflags="-H=windowsgui" -o bin/interactive.exe cmd/interactive/main.go

echo   Building analyze-cross.exe...
go build -ldflags="-H=windowsgui" -o bin/analyze-cross.exe cmd/analyze-cross/main.go

echo.
echo ========================================
echo Build complete. Output: bin\
echo ========================================
dir bin\*.exe
pause
