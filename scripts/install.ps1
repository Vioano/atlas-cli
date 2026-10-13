$ErrorActionPreference = 'Stop'
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
if (-not [Environment]::Is64BitOperatingSystem) { throw 'Atlas requires 64-bit Windows.' }
$atlasDirectory = Join-Path $env:LOCALAPPDATA 'Atlas'
New-Item -ItemType Directory -Force -Path $atlasDirectory | Out-Null
$atlasTemporary = Join-Path $atlasDirectory 'atlas-download.zip'
$atlasExecutable = Join-Path $atlasDirectory 'atlas.exe'
$atlasURL = 'https://github.com/Vioano/atlas-cli/releases/download/v0.2.11/atlas-windows-amd64.zip'
if ($env:ATLAS_DOWNLOAD_BASE) { $atlasURL = $env:ATLAS_DOWNLOAD_BASE.TrimEnd('/') + '/atlas-windows-amd64.zip' }
Invoke-WebRequest -UseBasicParsing -Uri $atlasURL -OutFile $atlasTemporary -Headers @{'User-Agent'='Atlas-CLI/0.2.11'}
Expand-Archive -Force -Path $atlasTemporary -DestinationPath $atlasDirectory
& $atlasExecutable version
if ($LASTEXITCODE -ne 0) { throw 'Downloaded Atlas did not run.' }
Remove-Item $atlasTemporary
$atlasUserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (($atlasUserPath -split ';') -notcontains $atlasDirectory) {
    [Environment]::SetEnvironmentVariable('Path', ($atlasDirectory + ';' + $atlasUserPath), 'User')
}
$env:Path = $atlasDirectory + ';' + $env:Path
Write-Output ('Installed ' + $atlasExecutable)
