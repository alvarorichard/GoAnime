; GoAnime Windows Installer Script
; This script is designed to be run from CI with files staged in the build directory

#define MyAppName "GoAnime"
#define MyAppVersion "1.6"
#define MyAppPublisher "GoAnime Team"
#define MyAppURL "https://github.com/alvarorichard/GoAnime"
#define MyAppExeName "goanime.exe"

[Setup]
AppId={{A1B2C3D4-E5F6-7890-ABCD-EF1234567890}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}/releases
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
AllowNoIcons=yes
; Output directory is relative to the script location
OutputDir=..\dist
OutputBaseFilename=GoAnime-Installer-{#MyAppVersion}
; SetupIconFile=..\assets\icon.ico  ; Uncomment if icon.ico is available
UninstallDisplayIcon={app}\{#MyAppExeName}
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=admin
ArchitecturesInstallIn64BitMode=x64compatible

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "addtopath"; Description: "Add GoAnime and MPV to PATH"; GroupDescription: "System Integration:"; Flags: checkedonce

[Files]
; Main application binary (staged in build/staging directory)
Source: "staging\goanime.exe"; DestDir: "{app}"; Flags: ignoreversion

; MPV binary and required DLLs for video playback
Source: "staging\bin\mpv.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
Source: "staging\bin\*.dll"; DestDir: "{app}\bin"; Flags: ignoreversion skipifsourcedoesntexist

[Icons]
; Start Menu shortcuts
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\{cm:UninstallProgram,{#MyAppName}}"; Filename: "{uninstallexe}"

; Desktop shortcut (optional)
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Run]
; Add to user PATH using setx for immediate effect (works better in Windows Sandbox)
; This runs before the postinstall option so PATH is ready when GoAnime starts
Filename: "{cmd}"; Parameters: "/C setx PATH ""%PATH%;{app};{app}\bin"""; Flags: runhidden runascurrentuser; Tasks: addtopath
; Option to run GoAnime after installation
Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent shellexec

[Registry]
; Note: We use setx in [Run] section for immediate PATH update (works in Windows Sandbox)
; Registry entries below are kept as backup for system-wide persistence
; They will be applied on next login/restart if setx fails
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; Tasks: addtopath; Check: NeedsAddPath('{app}')
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}\bin"; Tasks: addtopath; Check: NeedsAddPath('{app}\bin')

[Code]
// Windows API constant for broadcasting environment changes
const
  SMTO_ABORTIFHUNG = 2;
  WM_SETTINGCHANGE = $001A;
  HWND_BROADCAST = $FFFF;

// Import Windows API function to broadcast environment changes
function SendMessageTimeout(hWnd: HWND; Msg: UINT; wParam: WPARAM; lParam: PAnsiChar; 
  fuFlags, uTimeout: UINT; var lpdwResult: DWORD): UINT;
  external 'SendMessageTimeoutA@user32.dll stdcall';

// Broadcast environment change to all windows so PATH is updated immediately
procedure RefreshEnvironment;
var
  Res: DWORD;
begin
  // Notify all windows that environment variables have changed
  // This makes the PATH update take effect immediately without requiring logout/restart
  SendMessageTimeout(HWND_BROADCAST, WM_SETTINGCHANGE, 0, 'Environment', SMTO_ABORTIFHUNG, 5000, Res);
end;

function NeedsAddPath(Param: string): boolean;
var
  OrigPath: string;
begin
  if not RegQueryStringValue(HKLM, 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment', 'Path', OrigPath) then
  begin
    Result := True;
    exit;
  end;
  Result := Pos(';' + Param + ';', ';' + OrigPath + ';') = 0;
end;

procedure RemovePath(PathToRemove: string);
var
  OrigPath: string;
  NewPath: string;
  P: Integer;
begin
  if RegQueryStringValue(HKLM, 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment', 'Path', OrigPath) then
  begin
    NewPath := OrigPath;
    // Try to remove with semicolon before
    P := Pos(';' + PathToRemove, NewPath);
    if P > 0 then
    begin
      Delete(NewPath, P, Length(PathToRemove) + 1);
    end
    else
    begin
      // Try to remove with semicolon after
      P := Pos(PathToRemove + ';', NewPath);
      if P > 0 then
        Delete(NewPath, P, Length(PathToRemove) + 1);
    end;
    
    if NewPath <> OrigPath then
      RegWriteStringValue(HKLM, 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment', 'Path', NewPath);
  end;
end;

// Called after installation steps complete
procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssPostInstall then
  begin
    // Broadcast environment change so new PATH is available immediately
    // This is crucial for Windows Sandbox and fresh installs
    RefreshEnvironment;
  end;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usPostUninstall then
  begin
    // Remove paths added during installation
    RemovePath(ExpandConstant('{app}'));
    RemovePath(ExpandConstant('{app}\bin'));
    // Broadcast environment change after removing paths
    RefreshEnvironment;
  end;
end;
