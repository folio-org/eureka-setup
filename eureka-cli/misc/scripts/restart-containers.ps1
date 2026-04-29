# Script to restart all Docker containers with "eureka-" prefix
Write-Host "Finding containers with 'eureka-' prefix..." -ForegroundColor Cyan

# Get all container IDs that start with "eureka-"
$containers = docker ps -a --filter "name=^eureka-" --format "{{.ID}}"

if ([string]::IsNullOrWhiteSpace($containers)) {
  Write-Host "No containers found with 'eureka-' prefix." -ForegroundColor Yellow
  exit 0
}

Write-Host "Found the following containers:" -ForegroundColor Green
docker ps -a --filter "name=^eureka-" --format "table {{.ID}}`t{{.Names}}`t{{.Status}}"

Write-Host ""
Write-Host "Restarting containers (2 at a time in parallel)..." -ForegroundColor Cyan

# Convert container list to array
$containerArray = $containers -split "`n" | Where-Object { $_.Trim() -ne "" }
$total = $containerArray.Count
$index = 0

# Restart containers 2 at a time
while ($index -lt $total) {
  $jobs = @()
  
  # Start up to 2 container restarts in background
  for ($i = 0; $i -lt 2; $i++) {
    if ($index -lt $total) {
      $container = $containerArray[$index]
      $containerName = (docker inspect --format='{{.Name}}' $container) -replace '^/', ''
      Write-Host "Restarting $containerName ($container)..." -ForegroundColor Yellow
      
      # Start restart as background job
      $job = Start-Job -ScriptBlock {
        param($containerId)
        docker restart $containerId
      } -ArgumentList $container
      
      $jobs += $job
      $index++
    }
  }
  
  # Wait for all restarts in this batch to complete
  if ($jobs.Count -gt 0) {
    $jobs | Wait-Job | Remove-Job
  }
}

Write-Host ""
Write-Host "All eureka- containers have been restarted successfully" -ForegroundColor Green
Write-Host "Waiting 3 minutes for containers to stabilize..." -ForegroundColor Cyan
Start-Sleep -Seconds 180
Write-Host "Wait complete" -ForegroundColor Green
