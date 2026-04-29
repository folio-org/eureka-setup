# Script to add Eureka hostnames to Windows hosts file
# This script is idempotent - it can be run multiple times safely
# Usage: Run as Administrator in PowerShell

#Requires -RunAsAdministrator

# Define hostnames to add
$Hostnames = @(
  "postgres.eureka",
  "kafka.eureka",
  "vault.eureka",
  "keycloak.eureka",
  "kong.eureka"
)

$HostsFile = "$env:SystemRoot\System32\drivers\etc\hosts"
$IpAddress = "127.0.0.1"

Write-Host "Eureka Hosts Configuration Script" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan
Write-Host ""

# Verify hosts file exists
if (-not (Test-Path $HostsFile)) {
  Write-Host "Error: Hosts file not found at $HostsFile" -ForegroundColor Red
  exit 1
}

# Create backup of hosts file
$BackupFile = "$HostsFile.backup.$(Get-Date -Format 'yyyyMMdd_HHmmss')"
try {
  Copy-Item -Path $HostsFile -Destination $BackupFile -Force
  Write-Host "Created backup: $BackupFile" -ForegroundColor Green
  Write-Host ""
}
catch {
  Write-Host "Error: Failed to create backup - $_" -ForegroundColor Red
  exit 1
}

# Read current hosts file content
try {
  $HostsContent = Get-Content $HostsFile -ErrorAction Stop
}
catch {
  Write-Host "Error: Failed to read hosts file - $_" -ForegroundColor Red
  exit 1
}

# Track changes
$AddedCount = 0
$SkippedCount = 0
$EntriesToAdd = @()

# Process each hostname
foreach ($hostname in $Hostnames) {
  # Check if hostname already exists in hosts file
  $Pattern = "\s+$([regex]::Escape($hostname))(\s|$)"
  $Exists = $HostsContent | Where-Object { $_ -match $Pattern }
  
  if ($Exists) {
    Write-Host "Skipped: $hostname (already exists)" -ForegroundColor Yellow
    $SkippedCount++
  }
  else {
    $EntriesToAdd += "$IpAddress $hostname"
    Write-Host "Added: $IpAddress $hostname" -ForegroundColor Green
    $AddedCount++
  }
}

# Add new entries to hosts file if any
if ($EntriesToAdd.Count -gt 0) {
  try {
    # Ensure file ends with newline before appending
    $LastLine = $HostsContent[-1]
    if ($LastLine -and -not [string]::IsNullOrWhiteSpace($LastLine)) {
      Add-Content -Path $HostsFile -Value "" -NoNewline -ErrorAction Stop
    }
    
    # Add new entries
    Add-Content -Path $HostsFile -Value $EntriesToAdd -ErrorAction Stop
  }
  catch {
    Write-Host "Error: Failed to update hosts file - $_" -ForegroundColor Red
    Write-Host "Attempting to restore backup..." -ForegroundColor Yellow
    Copy-Item -Path $BackupFile -Destination $HostsFile -Force
    exit 1
  }
}

Write-Host ""
Write-Host "==================================" -ForegroundColor Cyan
Write-Host "Summary:" -ForegroundColor Cyan
Write-Host "  Added: $AddedCount" -ForegroundColor White
Write-Host "  Skipped: $SkippedCount" -ForegroundColor White
Write-Host "  Backup: $BackupFile" -ForegroundColor White
Write-Host ""

if ($AddedCount -gt 0) {
  Write-Host "Hosts file updated successfully!" -ForegroundColor Green
}
else {
  Write-Host "No changes needed - all hostnames already configured" -ForegroundColor Yellow
}

# Verify entries
Write-Host ""
Write-Host "Current Eureka entries in hosts file:" -ForegroundColor Cyan
Write-Host "-----------------------------------" -ForegroundColor Cyan
$UpdatedHostsContent = Get-Content $HostsFile
foreach ($hostname in $Hostnames) {
  $Pattern = "\s+$([regex]::Escape($hostname))(\s|$)"
  $Entry = $UpdatedHostsContent | Where-Object { $_ -match $Pattern }
  if ($Entry) {
    Write-Host $Entry -ForegroundColor White
  }
  else {
    Write-Host "(not found: $hostname)" -ForegroundColor Red
  }
}

Write-Host ""
Write-Host "Note: DNS cache flush may be required for changes to take effect immediately." -ForegroundColor Yellow
Write-Host "Run: ipconfig /flushdns" -ForegroundColor Yellow
