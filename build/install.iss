[Setup]
AppName=GoAnime Installer
AppVersion=1.3
DefaultDirName={pf}\GoAnime
DefaultGroupName=GoAnime
AllowNoIcons=yes
OutputBaseFilename=GoAnimeInstaller
UsePreviousAppDir=yes
Compression=lzma2
SolidCompression=yes

[Tasks]
Name: "desktopicon"; Description: "Create a &desktop shortcut"; GroupDescription: "Additional Options";

[Files]
; Copia o executável principal do GoAnime
Source: "C:\Users\krone\Documents\codes\GoAnime\build\goanime.exe"; DestDir: "{app}"; Flags: ignoreversion

; Copia os binários de mpv e yt-dlp para a pasta bin dentro do diretório de instalação
Source: "C:\Users\krone\Documents\codes\GoAnime\build\mpv.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
;Source: "C:\Users\krone\Documents\codes\GoAnime\build\yt-dlp.exe"; DestDir: "{app}\bin"; Flags: ignoreversion

[Icons]
; Cria o atalho no Menu Iniciar
Name: "{group}\GoAnime"; Filename: "{app}\goanime.exe"

; Cria o atalho na área de trabalho, se o usuário selecionar essa opção
Name: "{userdesktop}\GoAnime"; Filename: "{app}\goanime.exe"; Tasks: desktopicon

[Run]
; Adiciona mpv e yt-dlp ao PATH do usuário
Filename: "{cmd}"; Parameters: "/C setx PATH ""{app}\bin;%PATH%"""; Flags: runhidden runascurrentuser
