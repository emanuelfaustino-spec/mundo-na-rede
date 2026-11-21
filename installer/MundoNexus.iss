; Instalador por usuário: o atualizador interno pode substituir o executável
; sem UAC, e o atalho permanece estável no Menu Iniciar.
#define AppName "Mundo Nexus"
#ifndef AppVersion
  #define AppVersion "0.6.4"
#endif
#define AppPublisher "Mundo Nexus"
#define AppExeName "Mundo-Nexus.exe"
#ifndef SourceExe
  #define SourceExe "..\dist\Mundo-Nexus.exe"
#endif

[Setup]
; New identity: the former Mundo na Rede installer remembered the old folder.
; Keeping a distinct ID lets Mundo Nexus be installed cleanly in its own path.
AppId={{F894B65B-E3B1-4CF2-A2B7-E8AD2A2DB874}
AppName={#AppName}
AppVersion={#AppVersion}
AppPublisher={#AppPublisher}
DefaultDirName={localappdata}\Programs\Mundo Nexus
DefaultGroupName=Mundo Nexus
DisableProgramGroupPage=yes
PrivilegesRequired=lowest
OutputDir=..\dist\installer
OutputBaseFilename=Mundo-Nexus-Setup-v{#AppVersion}
SetupIconFile=..\ui\icon-nexus.ico
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

[InstallDelete]
; Remove the Start menu entry used by versions released before the rename.
Type: files; Name: "{userprograms}\Mundo na Rede.lnk"

[Icons]
Name: "{autoprograms}\Mundo Nexus"; Filename: "{app}\{#AppExeName}"; WorkingDir: "{app}"; IconFilename: "{app}\{#AppExeName}"
Name: "{autodesktop}\Mundo Nexus"; Filename: "{app}\{#AppExeName}"; WorkingDir: "{app}"; IconFilename: "{app}\{#AppExeName}"; Tasks: desktopicon

[Run]
Filename: "{app}\{#AppExeName}"; Description: "Abrir o Mundo Nexus agora"; Flags: nowait postinstall skipifsilent
