; GoAnime Windows Installer Script
; This script is designed to be run from CI with files staged in the build directory

#define MyAppName "GoAnime"
#define MyAppVersion "2.0.0"
#define MyAppPublisher "GoAnime Team"
#define MyAppURL "https://github.com/alvarorichard/GoAnime"
#define MyAppExeName "goanime.exe"
#define MyAppId "A1B2C3D4-E5F6-7890-ABCD-EF1234567890"

[Setup]
AppId={{{#MyAppId}}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}/releases
; Per-user install: no administrator rights, no UAC prompt, and the updater
; can replace the executable without elevation.
DefaultDirName={localappdata}\Programs\{#MyAppName}
DefaultGroupName={#MyAppName}
AllowNoIcons=yes
; Output directory is relative to the script location
OutputDir=..\dist
OutputBaseFilename=GoAnime-Installer-{#MyAppVersion}
SetupIconFile=..\docs\assets\logo.ico
UninstallDisplayIcon={app}\{#MyAppExeName}
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=lowest
; Notify Explorer after changing HKCU\Environment so a new terminal sees PATH.
ChangesEnvironment=yes
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"
Name: "addtopath"; Description: "Add GoAnime and MPV to PATH"; GroupDescription: "User Integration:"; Flags: checkedonce

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
; Option to run GoAnime after installation. The installer is already running
; without elevation, so this simply starts the app as the same user.
Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent

[Code]
// Check if a path is already in the current user's PATH. Windows treats PATH
// entries case-insensitively, so this comparison does too.
function NeedsAddUserPath(Param: string): boolean;
var
  OrigPath: string;
begin
  if not RegQueryStringValue(HKCU, 'Environment', 'Path', OrigPath) then
  begin
    Result := True;
    exit;
  end;
  Result := Pos(Lowercase(';' + Param + ';'), Lowercase(';' + OrigPath + ';')) = 0;
end;

// Add one path to the current user's PATH, preserving every existing entry.
// This is done in [Code] instead of [Registry] so an empty PATH does not gain
// a leading semicolon and two entries cannot overwrite each other.
procedure AddUserPath(PathToAdd: string);
var
  OrigPath: string;
  NewPath: string;
begin
  if not NeedsAddUserPath(PathToAdd) then
    exit;

  if not RegQueryStringValue(HKCU, 'Environment', 'Path', OrigPath) then
    OrigPath := '';

  if OrigPath = '' then
    NewPath := PathToAdd
  else if OrigPath[Length(OrigPath)] = ';' then
    NewPath := OrigPath + PathToAdd
  else
    NewPath := OrigPath + ';' + PathToAdd;

  RegWriteExpandStringValue(HKCU, 'Environment', 'Path', NewPath);
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if (CurStep = ssPostInstall) and WizardIsTaskSelected('addtopath') then
  begin
    AddUserPath(ExpandConstant('{app}'));
    AddUserPath(ExpandConstant('{app}\bin'));
  end;
end;

// Remove one path from the current user's PATH without touching any other
// entry. Called only from the uninstaller. Splitting and rejoining avoids the
// off-by-one of a Pos/Delete loop: the leading ';' added for the first entry
// shifts the index, which used to leave a stray ';' behind. Empty entries in
// the middle are preserved; a trailing separator is normalized away.
procedure RemoveUserPath(PathToRemove: string);
var
  OrigPath: string;
  Parts: TArrayOfString;
  Kept: TArrayOfString;
  I, Count: Integer;
  NewPath: string;
begin
  if not RegQueryStringValue(HKCU, 'Environment', 'Path', OrigPath) then
    exit;

  Parts := StringSplit(OrigPath, [';'], stAll);
  SetArrayLength(Kept, GetArrayLength(Parts));
  Count := 0;
  for I := 0 to GetArrayLength(Parts) - 1 do
  begin
    if not SameText(Parts[I], PathToRemove) then
    begin
      Kept[Count] := Parts[I];
      Count := Count + 1;
    end;
  end;
  SetArrayLength(Kept, Count);
  NewPath := StringJoin(';', Kept);

  if NewPath <> OrigPath then
    RegWriteExpandStringValue(HKCU, 'Environment', 'Path', NewPath);
end;

// Version 1.x installed for all users under Program Files. Its HKLM PATH entry
// precedes the user PATH, so leaving it in place would make Windows keep
// launching the old binary after this per-user install. Ask for a clean
// uninstall instead of creating a second, conflicting installation.
function IsOldPerMachineInstall(): Boolean;
var
  UninstallString: string;
  Key: string;
begin
  // AppId is written as {GUID}; the uninstall key is {GUID}_is1. The double
  // braces keep the literal braces around the expanded GUID.
  Key := 'Software\Microsoft\Windows\CurrentVersion\Uninstall\{{#MyAppId}}_is1';
  Result :=
    RegQueryStringValue(HKLM, Key, 'UninstallString', UninstallString) or
    RegQueryStringValue(HKLM32, Key, 'UninstallString', UninstallString);
  // A 64-bit install (the 1.x installer used x64compatible) keeps its
  // uninstall key in the 64-bit view. HKLM above is the 32-bit view in this
  // script, so check HKLM64 explicitly. It is only valid on 64-bit Windows.
  if (not Result) and IsWin64 then
    Result := RegQueryStringValue(HKLM64, Key, 'UninstallString', UninstallString);
end;

function InitializeSetup(): Boolean;
begin
  Result := not IsOldPerMachineInstall();
  if not Result then
    // SuppressibleMsgBox, not MsgBox: /SUPPRESSMSGBOXES must not wait for a
    // human in silent installs and in CI. It returns the default button.
    SuppressibleMsgBox('An older all-users installation of GoAnime was found.' + #13#10 + #13#10 +
      'Please uninstall it first (Settings > Apps > Installed apps), then run this installer again. ' +
      'This keeps a single GoAnime in PATH and avoids launching the old version.',
      mbError, MB_OK, IDOK);
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usPostUninstall then
  begin
    RemoveUserPath(ExpandConstant('{app}'));
    RemoveUserPath(ExpandConstant('{app}\bin'));
  end;
end;
