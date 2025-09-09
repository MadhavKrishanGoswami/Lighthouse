
@echo off
:: --- 1. Check for Administrator privileges ---
net session >nul 2>&1
if %errorLevel% == 0 (
    echo Administrator permissions confirmed.
) else (
    echo This installer must be run as an Administrator.
    pause
    exit /b
)

echo --- Lighthouse Host Agent Installer ---

:: --- 2. Set the orchestrator value by asking the user ---
set /p ORCHESTRATOR_URL="Enter the Orchestrator URL (e.g., 192.168.1.100:50051): "
if "%ORCHESTRATOR_URL%"=="" (
    echo Orchestrator URL cannot be empty. Aborting.
    pause
    exit /b
)

echo Configuration received. Installing...

:: --- 3. Create the configuration directory and file ---
set CONFIG_DIR="%ProgramData%\LighthouseHostAgent"
if not exist %CONFIG_DIR% mkdir %CONFIG_DIR%
echo # Configuration for the Lighthouse Host Agent > %CONFIG_DIR%\config.yaml
echo orchestrator_addr: "%ORCHESTRATOR_URL%" >> %CONFIG_DIR%\config.yaml

:: --- 4. Install the binary (the .exe file) ---
:: Assumes 'host-agent.exe' is in the same folder as this script.
set INSTALL_DIR="%ProgramFiles%\LighthouseHostAgent"
if not exist %INSTALL_DIR% mkdir %INSTALL_DIR%
copy host-agent.exe %INSTALL_DIR%\lighthouse-agent.exe

:: --- 5. "Open" the agent by installing and starting it as a service ---
echo Registering and starting the system service...
%INSTALL_DIR%\lighthouse-agent.exe install
%INSTALL_DIR%\lighthouse-agent.exe start

echo.
echo Lighthouse Host Agent installation complete!
echo The service is now running in the background.
pause
