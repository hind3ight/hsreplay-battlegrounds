@echo off
chcp 65001 >nul
echo ========================================
echo Battlegrounds Advisor - 一键安装
echo ========================================
echo.

:: 检查 Python
where py >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [错误] 未找到 Python，请从 https://python.org 安装
    pause
    exit /b 1
)

for /f "tokens=*" %%i in ('py --version') do set PYVER=%%i
echo [信息] 检测到 %PYVER%

:: 创建虚拟环境
if not exist ".venv" (
    echo.
    echo [步骤 1/3] 创建虚拟环境...
    py -m venv .venv
    echo [完成]
) else (
    echo.
    echo [步骤 1/3] 虚拟环境已存在，跳过
)

:: 安装依赖
echo.
echo [步骤 2/3] 安装依赖...
set PYTHONIOENCODING=utf-8
.venv\Scripts\pip install --upgrade pip
.venv\Scripts\pip install -r reader\requirements.txt
echo [完成]

:: 运行测试
echo.
echo [步骤 3/3] 运行 Mock 测试...
.venv\Scripts\python tests\test_mock.py

echo.
echo ========================================
echo 安装与测试完成！
echo ========================================
pause
