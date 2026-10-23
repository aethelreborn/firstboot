$ErrorActionPreference = 'Stop'
$repo = 'aethelreborn/firstboot'
$arch = if ([Environment]::Is64BitOperatingSystem) { 'amd64' } else { throw 'Unsupported architecture' }
$release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
$version = $release.tag_name
if (-not $version) { throw 'No release found; build from source with Go instead.' }
$tmp = Join-Path $env:TEMP ('firstboot-' + [Guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null
$url = "https://github.com/$repo/releases/download/$version/firstboot_windows_$arch.zip"
Invoke-WebRequest $url -OutFile (Join-Path $tmp 'firstboot.zip')
Expand-Archive (Join-Path $tmp 'firstboot.zip') -DestinationPath $tmp
$bin = Join-Path $env:LOCALAPPDATA 'firstboot'
New-Item -ItemType Directory -Force -Path $bin | Out-Null
Copy-Item (Join-Path $tmp 'firstboot.exe') (Join-Path $bin 'firstboot.exe') -Force
Write-Host "Installed firstboot to $bin\firstboot.exe"
