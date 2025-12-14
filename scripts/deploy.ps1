param (
    [string]$ServerIP = "192.168.0.167",
    [string]$Username = "derlin",
    [string]$RemoteDir = "/opt/wxbot",
    [string]$IdentityFile = "",
    [string]$Service = ""
)

$ErrorActionPreference = "Stop"
$TempZip = "/tmp/botmatrix_deploy.zip"
$LocalZip = "botmatrix_deploy.zip"

Write-Host "========================================" -ForegroundColor Cyan
if ($Service) {
    Write-Host "   Deploying Service [$Service] to ${Username}@${ServerIP}" -ForegroundColor Cyan
} else {
    Write-Host "   Deploying ALL Services to ${Username}@${ServerIP}" -ForegroundColor Cyan
}
Write-Host "========================================" -ForegroundColor Cyan

# 1. Pack
Write-Host "[Step 1/3] Packing project..." -ForegroundColor Green
python scripts/pack_project.py

if (-not (Test-Path $LocalZip)) {
    Write-Error "Error: $LocalZip not found!"
    exit 1
}

# 2. Upload
Write-Host "[Step 2/3] Uploading to server..." -ForegroundColor Green
$scpArgs = @()
if ($IdentityFile) {
    $scpArgs += "-i", $IdentityFile
}
$scpArgs += $LocalZip, "${Username}@${ServerIP}:${TempZip}"

Write-Host "Running: scp $scpArgs"
scp @scpArgs

if ($LASTEXITCODE -ne 0) {
    Write-Error "SCP failed with exit code $LASTEXITCODE"
    exit 1
}

# 3. Deploy
Write-Host "[Step 3/3] Executing remote commands..." -ForegroundColor Green

# Construct the docker command based on whether a specific service is requested
$DockerCmd = ""
if ($Service) {
    # If a service is specified, we only update that service (and its deps if needed, but usually we want to isolate)
    # We do NOT run 'down' to keep other services running
    $DockerCmd = "docker-compose up -d --build --no-deps $Service"
} else {
    # Full deployment: down everything and bring it back up
    $DockerCmd = "docker-compose down --remove-orphans && docker-compose up -d --build"
}

$remoteCommands = @"
    echo '--> Creating directory...'
    mkdir -p ${RemoteDir}
    
    echo '--> Unzipping...'
    unzip -o ${TempZip} -d ${RemoteDir}
    rm ${TempZip}
    
    echo '--> Switching directory...'
    cd ${RemoteDir}
    
    echo '--> executing: ${DockerCmd}'
    ${DockerCmd}
    
    echo '--> Deployment SUCCESS!'
"@

$sshArgs = @()
if ($IdentityFile) {
    $sshArgs += "-i", $IdentityFile
}
$sshArgs += "-t", "${Username}@${ServerIP}", $remoteCommands

Write-Host "Running: ssh $sshArgs"
ssh @sshArgs

Write-Host "Done." -ForegroundColor Cyan
