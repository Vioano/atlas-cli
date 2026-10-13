$ErrorActionPreference = 'Stop'
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
if (-not [Environment]::Is64BitOperatingSystem) { throw 'Atlas requires 64-bit Windows.' }
$atlasDirectory = Join-Path $env:LOCALAPPDATA 'Atlas'
New-Item -ItemType Directory -Force -Path $atlasDirectory | Out-Null
$atlasTemporary = Join-Path $atlasDirectory 'atlas-download.exe'
$atlasExecutable = Join-Path $atlasDirectory 'atlas.exe'
$atlasURL = 'https://github.com/Vioano/atlas-cli/releases/download/v0.1.0/atlas-windows-amd64.exe'
if ($env:ATLAS_DOWNLOAD_BASE) { $atlasURL = $env:ATLAS_DOWNLOAD_BASE.TrimEnd('/') + '/atlas-windows-amd64.exe' }
Invoke-WebRequest -UseBasicParsing -Uri $atlasURL -OutFile $atlasTemporary -Headers @{'User-Agent'='Atlas-CLI/0.1.0'}
& $atlasTemporary version
if ($LASTEXITCODE -ne 0) { Remove-Item $atlasTemporary; throw 'Downloaded Atlas did not run.' }
Move-Item -Force $atlasTemporary $atlasExecutable
$atlasUserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (($atlasUserPath -split ';') -notcontains $atlasDirectory) {
    [Environment]::SetEnvironmentVariable('Path', ($atlasDirectory + ';' + $atlasUserPath), 'User')
}
$env:Path = $atlasDirectory + ';' + $env:Path
Write-Output ('Installed ' + $atlasExecutable)
