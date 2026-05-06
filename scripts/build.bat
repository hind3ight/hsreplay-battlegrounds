@echo off
chcp 65001 >nul
echo ========================================
echo Battlegrounds Advisor - Go Build Script
echo ========================================
echo.

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
echo [步骤 1/2] 安装 Go 依赖...
go mod download
echo [完成]
echo.

:: 编译 Windows amd64 二进制
echo [步骤 2/2] 编译 Windows 二进制...
mkdir bin 2>nul

set GOOS=windows
set GOARCH=amd64

echo   Building fetch...
go build -o bin/fetch.exe cmd/fetch/main.go

echo   Building analyze...
go build -o bin/analyze.exe cmd/analyze/main.go

echo   Building interactive...
go build -o bin/interactive.exe cmd/interactive/main.go

echo   Building analyze-cross...
go build -o bin/analyze-cross.exe cmd/analyze-cross/main.go

echo.
echo ========================================
echo 编译完成！输出目录: bin\
echo ========================================
dir bin\*.exe
pause
