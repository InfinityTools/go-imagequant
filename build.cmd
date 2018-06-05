@echo off

REM A build script that automatically picks the right library from the subfolders in "libs".
REM Use this script if you are unable or don't want to use the system library.

setlocal

REM Checking Go compiler
where /q go || (
  echo Go compiler not found.
  goto Failed
)

REM Package-specific settings
set pkgRoot=github.com/InfinityTools
set ldargs=-limagequant -lm
set uselibdir=0

REM Evaluating command line parameters
:ArgsLoop
if "%~1"=="" goto ArgsFinished

if "%~1"=="--libdir" (
  set uselibdir=1
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

REM Handling overridden libdir
if /i %uselibdir% EQU 1 (
  if not exist "%libdir:/=\%" (
    echo Directory does not exist: %libdir%
    goto Failed
  )
)
if /i %uselibdir% EQU 1 (
  echo Using libdir: %libdir%
)

REM Autodetect libdir
if /i %uselibdir% EQU 0 (
  for /f "tokens=* usebackq" %%a in (`go env GOOS`) do (
    set libos=%%a
  )
)
if /i %uselibdir% EQU 0 (
  for /f "tokens=* usebackq" %%a in (`go env GOARCH`) do (
    set libarch=%%a
  )
)
if /i %uselibdir% EQU 0 (
  set pkgImagequant=%pkgRoot%/go-imagequant
)
if /i %uselibdir% EQU 0 (
  go list %pkgImagequant% >nul 2>&1 || (
    echo Package not found: %pkgImagequant%
    goto Failed
  )
)
if /i %uselibdir% EQU 0 (
  for /f "tokens=* usebackq" %%a in (`go list -f {{.Dir}} %pkgImagequant%`) do (
    set ldprefix=%%a
  )
)
if /i %uselibdir% EQU 0 (
  set ldprefix=%ldprefix:\=/%
)
if /i %uselibdir% EQU 0 (
  echo Detected: os=%libos%, arch=%libarch%
  set libdir=%ldprefix%/libs/%libos%/%libarch%
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
