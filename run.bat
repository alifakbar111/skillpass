@echo off
setlocal

:: Refresh PATH so docker, go, and bun are found
set "PATH=%PATH%;%ProgramFiles%\Docker\Docker\resources\bin;%ProgramFiles%\Go\bin;%USERPROFILE%\AppData\Local\bun"
for /f "tokens=*" %%i in ('powershell -NoProfile -Command "[System.Environment]::GetEnvironmentVariable(\"Path\",\"Machine\")"') do set "MACHINE_PATH=%%i"
for /f "tokens=*" %%i in ('powershell -NoProfile -Command "[System.Environment]::GetEnvironmentVariable(\"Path\",\"User\")"') do set "USER_PATH=%%i"
set "PATH=%MACHINE_PATH%;%USER_PATH%"

echo.
echo  SkillPass Dev Launcher
echo  ========================
echo.

:: 1. Start PostgreSQL via Docker
echo [1/3] Starting PostgreSQL...
docker compose up db -d
if errorlevel 1 (
    echo ERROR: Docker failed. Make sure Docker Desktop is running.
    pause
    exit /b 1
)

:: Wait a moment for Postgres to be ready
timeout /t 3 /nobreak >nul

:: 2. Run migrations
echo.
echo [2/3] Running migrations...
bun run db:migrate
if errorlevel 1 (
    echo ERROR: Migrations failed.
    pause
    exit /b 1
)

:: 3. Start dev servers
echo.
echo [3/3] Starting dev servers...
echo.
echo   Web  -^>  http://localhost:4200
echo   API  -^>  http://localhost:1234
echo.
bun run dev
