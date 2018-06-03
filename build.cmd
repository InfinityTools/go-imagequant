@echo off

REM A build script that automatically picks the right library from the subfolders in "libs".
REM Use this script if you are unable or don't want to use the system library.

setlocal

REM Package-specific libraries
set ldargs=-limagequant -lm

REM Checking Go compiler
where /q go || (
  echo Go compiler not found.
  goto Failed
)

REM Evaluating command line parameters
:ArgsLoop
if "%~1"=="" goto ArgsFinished

if "%~1"=="--libdir" (
  set libdir=%~2
  shift
  goto ArgsUpdate
)

if "%~1"=="--help" (
echo Usage: %~n0%~x0 [options]
echo.
echo Options:
echo   --libdir path    Override library path
echo   --help           This help
goto Finished
)

:ArgsUpdate
shift
goto ArgsLoop

:ArgsFinished

if [%libdir%]==[] (
  set customLibdir=0
) else (
  set customLibdir=1
)

REM Handling custom libdir
if /i %customLibdir% NEQ 0 (
  if not exist %libdir:/=\% (
    echo Directory does not exist: %libdir%
    goto Failed
  )
  echo Using libdir: %libdir%
)

REM Autodetect libdir
if /i %customLibdir% EQU 0 (
  for /f "tokens=* usebackq" %%a in (`go env GOOS`) do (
    set libos=%%a
  )
  for /f "tokens=* usebackq" %%a in (`go env GOARCH`) do (
    set libarch=%%a
  )
)
if /i %customLibdir% EQU 0 (
  echo Detected: os=%libos%, arch=%libarch%
  set libdir=libs/%libos%/%libarch%
)

echo Building library...
set CGO_LDFLAGS=-L%libdir% %ldargs%
go build && go install && goto Success || goto Failed

:Failed
echo Cancelled.
endlocal
exit /b 1

:Success
echo Finished.

:Finished
endlocal
