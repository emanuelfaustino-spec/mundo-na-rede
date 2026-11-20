param(
  [Parameter(Mandatory=$true)][string]$IdentityName,
  [Parameter(Mandatory=$true)][string]$Publisher,
  [Parameter(Mandatory=$true)][string]$PublisherDisplayName,
  [switch]$ValidationOnly
)
$ErrorActionPreference = 'Stop'
$root = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$versionSource = Get-Content (Join-Path $root 'version.go') -Raw
if ($versionSource -notmatch 'appVersion\s*=\s*"([0-9]+\.[0-9]+\.[0-9]+)"') { throw 'Versao nao encontrada em version.go' }
$version = "$($Matches[1]).0"
if (-not $ValidationOnly -and ($IdentityName -like '*Validation*' -or $Publisher -like '*Validation*')) { throw 'Identidade de validacao nao pode ser usada para o pacote da Store.' }
$sdk = 'C:\Program Files (x86)\Windows Kits\10\bin\10.0.26100.0\x64\makeappx.exe'
if (-not (Test-Path $sdk)) { throw 'makeappx.exe nao encontrado no Windows SDK.' }
$go = 'C:\Program Files\Go\bin\go.exe'
if (-not (Test-Path $go)) { throw 'Go nao encontrado.' }
$stage = Join-Path $root 'dist\store-staging'
$output = Join-Path $root ("dist\Mundo-Nexus-$version-x64" + $(if($ValidationOnly){'-VALIDATION-ONLY'}else{''}) + '.msix')
if (Test-Path $stage) { Remove-Item $stage -Force -Recurse }
New-Item -ItemType Directory -Force -Path (Join-Path $stage 'Assets') | Out-Null
Push-Location $root
try {
  & $go test ./...
  if ($LASTEXITCODE -ne 0) { throw 'Os testes falharam.' }
  & $go vet ./...
  if ($LASTEXITCODE -ne 0) { throw 'go vet falhou.' }
  & $go build -trimpath '-ldflags=-H=windowsgui -s -w -X main.distributionChannel=store' -o (Join-Path $stage 'Mundo-Nexus.exe') .
  if ($LASTEXITCODE -ne 0) { throw 'Compilacao falhou.' }
} finally { Pop-Location }
[xml]$xml = Get-Content (Join-Path $PSScriptRoot 'AppxManifest.template.xml') -Raw
$ns = New-Object System.Xml.XmlNamespaceManager($xml.NameTable)
$ns.AddNamespace('a','http://schemas.microsoft.com/appx/manifest/foundation/windows10')
$identity = $xml.SelectSingleNode('/a:Package/a:Identity',$ns)
$identity.SetAttribute('Name',$IdentityName)
$identity.SetAttribute('Publisher',$Publisher)
$identity.SetAttribute('Version',$version)
$publisherNode = $xml.SelectSingleNode('/a:Package/a:Properties/a:PublisherDisplayName',$ns)
$publisherNode.InnerText = $PublisherDisplayName
$xml.Save((Join-Path $stage 'AppxManifest.xml'))
Add-Type -AssemblyName System.Drawing
$icon = Join-Path $root 'ui\icon.png'
$source = [System.Drawing.Image]::FromFile($icon)
try {
  $sizes = @(
    @('StoreLogo.png',50,50),
    @('Square44x44Logo.png',44,44),
    @('Square150x150Logo.png',150,150),
    @('Wide310x150Logo.png',310,150),
    @('Square310x310Logo.png',310,310)
  )
  foreach ($item in $sizes) {
    $bitmap = New-Object System.Drawing.Bitmap ([int]$item[1]),([int]$item[2])
    try {
      $g = [System.Drawing.Graphics]::FromImage($bitmap)
      try {
        $g.Clear([System.Drawing.Color]::Transparent)
        $g.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
        $factor = [Math]::Min($bitmap.Width / $source.Width,$bitmap.Height / $source.Height)
        $w = [int][Math]::Round($source.Width * $factor)
        $h = [int][Math]::Round($source.Height * $factor)
        $g.DrawImage($source, [int](($bitmap.Width-$w)/2), [int](($bitmap.Height-$h)/2), $w, $h)
      } finally { $g.Dispose() }
      $bitmap.Save((Join-Path $stage ('Assets\'+$item[0])),[System.Drawing.Imaging.ImageFormat]::Png)
    } finally { $bitmap.Dispose() }
  }
} finally { $source.Dispose() }
& $sdk pack /d $stage /p $output /o
if ($LASTEXITCODE -ne 0) { throw 'Falha ao gerar pacote MSIX.' }
$sha = [System.Security.Cryptography.SHA256]::Create()
$stream = [System.IO.File]::OpenRead($output)
try { $hash = [BitConverter]::ToString($sha.ComputeHash($stream)).Replace('-','') } finally { $stream.Dispose(); $sha.Dispose() }
Write-Host "SHA256: $hash"
Write-Host "PACOTE: $output"
if ($ValidationOnly) { Write-Warning 'Pacote de validacao, com identidade ficticia. NAO publicar nem instalar como versao final.' }
else { Write-Warning 'Pacote sem assinatura local; submetê-lo apenas pela Microsoft Store apos conferir identidade e validar WACK.' }
