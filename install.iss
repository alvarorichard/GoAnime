[Setup]
AppName=YT-DLP and MPV Installer
AppVersion=1.0
DefaultDirName={pf}\YT-MPV
DefaultGroupName=YT-MPV
Compression=lzma
SolidCompression=yes
PrivilegesRequired=admin

[Files]
Source: "C:\Program Files\7-Zip\7z.exe"; DestDir: "{tmp}"; Flags: dontcopy
Source: "C:\Program Files\7-Zip\7z.dll"; DestDir: "{tmp}"; Flags: dontcopy

[Code]
const
  YTDLPURL = 'https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe';
  MPVURL = 'https://sourceforge.net/projects/mpv-player-windows/files/latest/download';
  WM_SETTINGCHANGE = $1A;
  SMTO_ABORTIFHUNG = 2;

type
  // Removed the duplicate type definitions of HWND, UINT, DWORD_PTR, and HRESULT
  // UINT = LongWord;
  WPARAM = LongWord;
  LPARAM = LongInt;
  // DWORD_PTR = LongWord; // Removed this as it is already predefined
  LRESULT = LongInt;
  // HRESULT = LongInt; // Removed this as it is already predefined
  DWORD = LongWord;

function SendMessageTimeout(hWnd: LongWord; Msg: LongWord; wParam: LongWord;
  lParam: LongInt; fuFlags: LongWord; uTimeout: LongWord; var lpdwResult: LongWord): LongInt;
  external 'SendMessageTimeoutW@user32.dll stdcall';

function URLDownloadToFile(Caller: Integer; URL: AnsiString; FileName: AnsiString; Reserved: Integer; StatusCB: Integer): LongInt;
  external 'URLDownloadToFileA@urlmon.dll stdcall';

function DownloadFile(const URL, FileName: string): Boolean;
begin
  Result := URLDownloadToFile(0, URL, FileName, 0, 0) = 0;
end;

function Unzip(SourceFile, DestDir: string): Boolean;
var
  ResultCode: Integer;
begin
  Result := Exec(ExpandConstant('{tmp}\7za.exe'), 'x "' + SourceFile + '" -o"' + DestDir + '" -y', '',
    SW_HIDE, ewWaitUntilTerminated, ResultCode) and (ResultCode = 0);
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  InstallDir: string;
  YTDLPPath: string;
  MPVArchivePath: string;
  MPVExtractDir: string;
begin
  if CurStep = ssInstall then
  begin
    InstallDir := ExpandConstant('{app}');
    YTDLPPath := InstallDir + '\yt-dlp.exe';
    MPVArchivePath := InstallDir + '\mpv.7z';
    MPVExtractDir := InstallDir;

    if not DirExists(InstallDir) then
    begin
      if not ForceDirectories(InstallDir) then
      begin
        MsgBox('Failed to create install directory: ' + InstallDir, mbError, MB_OK);
        WizardForm.Close;
        Exit;
      end;
    end;

    // Download yt-dlp
    if DownloadFile(YTDLPURL, YTDLPPath) then
    begin
      Log('yt-dlp.exe downloaded successfully.');
    end
    else
    begin
      MsgBox('Failed to download yt-dlp.exe', mbError, MB_OK);
      WizardForm.Close;
      Exit;
    end;

    // Download mpv archive
    if DownloadFile(MPVURL, MPVArchivePath) then
    begin
      Log('mpv archive downloaded successfully.');
    end
    else
    begin
      MsgBox('Failed to download mpv archive', mbError, MB_OK);
      WizardForm.Close;
      Exit;
    end;

    // Extract mpv
    ExtractTemporaryFile('7z.dll');
    ExtractTemporaryFile('7za.exe');
    if Unzip(MPVArchivePath, MPVExtractDir) then
    begin
      Log('mpv archive extracted successfully.');
    end
    else
    begin
      MsgBox('Failed to extract mpv archive', mbError, MB_OK);
      WizardForm.Close;
      Exit;
    end;

    // Now, add InstallDir to system PATH
    if not AddToPath(InstallDir) then
    begin
      MsgBox('Failed to update system PATH', mbError, MB_OK);
      WizardForm.Close;
      Exit;
    end;
  end;
end;

function AddToPath(NewPath: string): Boolean;
var
  OldPath, NewEnvPath: string;
  dwResult: LongWord;
begin
  Result := False;
  // Read the existing Path value
  if not RegQueryStringValue(HKLM, 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment', 'Path', OldPath) then
    OldPath := '';

  // Check if NewPath is already in OldPath
  if Pos(LowerCase(NewPath), LowerCase(OldPath)) = 0 then
  begin
    // Append NewPath to OldPath
    if (OldPath <> '') and (OldPath[Length(OldPath)] <> ';') then
      OldPath := OldPath + ';';
    NewEnvPath := OldPath + NewPath;

    // Write the new Path value
    if RegWriteStringValue(HKLM, 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment', 'Path', NewEnvPath) then
    begin
      // Notify the system about the environment variable change
      SendMessageTimeout(HWND_BROADCAST, WM_SETTINGCHANGE, 0,
        LPARAM(PChar('Environment')), SMTO_ABORTIFHUNG, 5000, dwResult);
      Result := True;
    end
    else
    begin
      // Failed to write the new Path value
      Result := False;
    end;
  end
  else
  begin
    // Path already contains the directory
    Result := True;
  end;
end;
