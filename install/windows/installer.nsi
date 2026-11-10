; LightLauncher installer (NSIS).
; Built by scripts/build-windows.sh. Manual build:
;   makensis -DVERSION=1.2.3 -DSTAGE_DIR=C:\path\to\stage -DOUTFILE=C:\path\setup.exe install/windows/installer.nsi

!ifndef VERSION
  !error "VERSION is not set. Pass -DVERSION=<version>."
!endif
!ifndef STAGE_DIR
  !error "STAGE_DIR is not set. Pass -DSTAGE_DIR=<folder with the built files>."
!endif
!ifndef OUTFILE
  !define OUTFILE "light-launcher-setup.exe"
!endif

!define APP_NAME "LightLauncher"
!define APP_EXE "light-launcher.exe"
!define DATA_DIR "light-launcher"
!define UNINSTALL_KEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}"
!define WEBVIEW2_CLIENT "Software\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}"
!define WEBVIEW2_PAGE "https://developer.microsoft.com/microsoft-edge/webview2/"
!define WEBVIEW2_SETUP_URL "https://go.microsoft.com/fwlink/p/?LinkId=2124703"

Name "${APP_NAME}"
OutFile "${OUTFILE}"
InstallDir "$PROGRAMFILES64\${APP_NAME}"
InstallDirRegKey HKLM "${UNINSTALL_KEY}" "InstallLocation"
RequestExecutionLevel admin
Unicode true
SetCompressor /SOLID lzma

!include "MUI2.nsh"
!include "LogicLib.nsh"

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Section "LightLauncher" SecMain
  SetShellVarContext all
  Call EnsureWebView2

  SetOutPath "$INSTDIR"
  File /r "${STAGE_DIR}\*.*"
  WriteUninstaller "$INSTDIR\uninstall.exe"

  CreateDirectory "$SMPROGRAMS\${APP_NAME}"
  CreateShortCut "$SMPROGRAMS\${APP_NAME}\${APP_NAME}.lnk" "$INSTDIR\${APP_EXE}"
  CreateShortCut "$SMPROGRAMS\${APP_NAME}\Uninstall ${APP_NAME}.lnk" "$INSTDIR\uninstall.exe"
  CreateShortCut "$DESKTOP\${APP_NAME}.lnk" "$INSTDIR\${APP_EXE}"

  WriteRegStr HKLM "${UNINSTALL_KEY}" "DisplayName" "${APP_NAME}"
  WriteRegStr HKLM "${UNINSTALL_KEY}" "DisplayVersion" "${VERSION}"
  WriteRegStr HKLM "${UNINSTALL_KEY}" "Publisher" "LightLauncher project"
  WriteRegStr HKLM "${UNINSTALL_KEY}" "InstallLocation" "$INSTDIR"
  WriteRegStr HKLM "${UNINSTALL_KEY}" "DisplayIcon" "$INSTDIR\${APP_EXE}"
  WriteRegStr HKLM "${UNINSTALL_KEY}" "UninstallString" '"$INSTDIR\uninstall.exe"'
  WriteRegDWORD HKLM "${UNINSTALL_KEY}" "NoModify" 1
  WriteRegDWORD HKLM "${UNINSTALL_KEY}" "NoRepair" 1
SectionEnd

; Sets $R0 to 1 when the WebView2 runtime is registered, otherwise 0.
Function HasWebView2
  Push $R0
  ReadRegStr $R0 HKLM "SOFTWARE\WOW6432Node\${WEBVIEW2_CLIENT}" "pv"
  ${If} $R0 == ""
    ReadRegStr $R0 HKLM "SOFTWARE\${WEBVIEW2_CLIENT}" "pv"
  ${EndIf}
  ${If} $R0 == ""
    ReadRegStr $R0 HKCU "Software\${WEBVIEW2_CLIENT}" "pv"
  ${EndIf}
  ${If} $R0 == ""
  ${OrIf} $R0 == "0.0.0.0"
    StrCpy $R0 0
  ${Else}
    StrCpy $R0 1
  ${EndIf}
  Exch $R0
FunctionEnd

; Offers to install the WebView2 runtime with Microsoft's bootstrapper.
Function EnsureWebView2
  Call HasWebView2
  Pop $R0
  ${If} $R0 == 1
    Return
  ${EndIf}

  MessageBox MB_YESNO|MB_ICONQUESTION "LightLauncher needs the Microsoft Edge WebView2 Runtime.$\r$\n$\r$\nInstall it now? This needs an internet connection.$\r$\n$\r$\nChoose No to install it later from:$\r$\n${WEBVIEW2_PAGE}" IDYES install_runtime
  Return

  install_runtime:
  ExecWait '"$SYSDIR\WindowsPowerShell\v1.0\powershell.exe" -NoProfile -ExecutionPolicy Bypass -Command "[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri ${WEBVIEW2_SETUP_URL} -OutFile $TEMP\MicrosoftEdgeWebview2Setup.exe"' $R0
  ${If} $R0 == 0
    ExecWait '"$TEMP\MicrosoftEdgeWebview2Setup.exe" /silent /install' $R0
  ${EndIf}
  Delete "$TEMP\MicrosoftEdgeWebview2Setup.exe"

  Call HasWebView2
  Pop $R0
  ${If} $R0 == 0
    MessageBox MB_OK|MB_ICONEXCLAMATION "The WebView2 Runtime could not be installed. LightLauncher will not start until it is.$\r$\n$\r$\n${WEBVIEW2_PAGE}"
  ${EndIf}
FunctionEnd

Section "Uninstall"
  SetShellVarContext all
  Delete "$SMPROGRAMS\${APP_NAME}\${APP_NAME}.lnk"
  Delete "$SMPROGRAMS\${APP_NAME}\Uninstall ${APP_NAME}.lnk"
  RMDir "$SMPROGRAMS\${APP_NAME}"
  Delete "$DESKTOP\${APP_NAME}.lnk"

  RMDir /r "$INSTDIR"

  DeleteRegKey HKLM "${UNINSTALL_KEY}"

  ; User data is kept unless the user asks for it to be removed.
  SetShellVarContext current
  MessageBox MB_YESNO|MB_ICONQUESTION|MB_DEFBUTTON2 "Also delete your LightLauncher data?$\r$\n$\r$\nThis removes your game settings, prefixes and logs in:$\r$\n$APPDATA\${DATA_DIR}$\r$\n$\r$\nChoose No to keep them." IDNO keep_data
  RMDir /r "$APPDATA\${DATA_DIR}"
  keep_data:
SectionEnd
