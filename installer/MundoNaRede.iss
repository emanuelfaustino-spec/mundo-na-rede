; Instalador por usuário: o atualizador interno pode substituir o executável
; sem UAC, e o atalho permanece estável no Menu Iniciar.
#define AppName "Mundo Nexo"
#ifndef AppVersion
  #define AppVersion "0.4.1"
#endif
#define AppPublisher "Mundo Nexo"
#define AppExeName "Mundo-Nexo.exe"
#ifndef SourceExe
  #define SourceExe "..\dist\Mundo-Nexo.exe"
#endif

[Setup]
AppId={{8CA2CE45-4CE7-4D90-B64F-39A0A36A373E}
AppName={#AppName}
AppVersion={#AppVersion}
AppPublisher={#AppPublisher}
DefaultDirName={localappdata}\Programs\Mundo Nexo
DefaultGroupName=Mundo Nexo
DisableProgramGroupPage=yes
PrivilegesRequired=lowest
OutputDir=C:\Users\emanu\Documents\Codex\2026-11-18\como-voc\outputs
OutputBaseFilename=Mundo-Nexo-Setup-v{#AppVersion}
SetupIconFile=..\ui\icon.ico
LicenseFile=termos-de-uso.txt
UninstallDisplayIcon={app}\{#AppExeName}
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
ArchitecturesInstallIn64BitMode=x64compatible

[Languages]
Name: "portuguese"; MessagesFile: "compiler:Languages\Portuguese.isl"

[Tasks]
Name: "desktopicon"; Description: "Criar atalho na área de trabalho"; GroupDescription: "Atalhos adicionais:"; Flags: unchecked

[Files]
Source: "{#SourceExe}"; DestDir: "{app}"; DestName: "{#AppExeName}"; Flags: ignoreversion

[Icons]
Name: "{autoprograms}\Mundo Nexo"; Filename: "{app}\{#AppExeName}"; WorkingDir: "{app}"; IconFilename: "{app}\{#AppExeName}"
Name: "{autodesktop}\Mundo Nexo"; Filename: "{app}\{#AppExeName}"; WorkingDir: "{app}"; IconFilename: "{app}\{#AppExeName}"; Tasks: desktopicon

[Run]
Filename: "{app}\{#AppExeName}"; Description: "Abrir o Mundo Nexo agora"; Flags: nowait postinstall skipifsilent
