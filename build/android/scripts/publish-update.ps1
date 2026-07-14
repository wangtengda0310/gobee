# publish-update.ps1 - After building APK, compute SHA256, generate latest.json, scp to itsnot.fun.
#
# Prerequisite: APK already built at build/android/app/build/outputs/apk/debug/app-debug.apk.
#   Build (in worktree root, git bash):
#     export PATH="/c/Users/<u>/go/bin:/c/Users/<u>/AppData/Local/Programs/Git/usr/bin:$PATH"
#     export ANDROID_HOME="D:/Android/Sdk" ANDROID_NDK_HOME="D:/Android/Sdk/ndk/26.3.11579264"
#     wails3 task android:compile:go:shared
#     cd build/android; JAVA_HOME="C:/Program Files/Android/Android Studio/jbr" cmd //c gradlew.bat assembleDebug
#
# Usage (increment versionCode in build.gradle BEFORE publishing):
#   pwsh build/android/scripts/publish-update.ps1 -Notes "fix xxx"
#
# NOTE: Keep this script ASCII-only. Windows PowerShell 5.1 reads BOM-less UTF-8 .ps1 as GBK,
#       non-ASCII comments/strings get corrupted and cause parse errors (same pitfall as cdp_eval.ps1).
param(
    [string]$Notes = "version update"
)
$ErrorActionPreference = "Stop"
$ROOT = (Resolve-Path "$PSScriptRoot/../../..").Path
$GRADLE = Join-Path $ROOT "build/android/app/build.gradle"
$APK = Join-Path $ROOT "build/android/app/build/outputs/apk/debug/app-debug.apk"
$REMOTE_DIR = "root@itsnot.fun:/root/itsnot.fun/nginx/html/rain-qa-func"

if (-not (Test-Path $APK)) {
    throw "APK not found: $APK. Build first (see header comments)."
}

# Read versionCode/versionName from build.gradle (should have been incremented before publishing)
$gradleContent = Get-Content $GRADLE -Raw
$vc = [int]([regex]::Match($gradleContent, 'versionCode\s+(\d+)').Groups[1].Value)
$vn = [regex]::Match($gradleContent, 'versionName\s+"([^"]+)"').Groups[1].Value
Write-Host "Publishing: versionCode=$vc versionName=$vn"

# SHA256 + filename
$sha = (Get-FileHash $APK -Algorithm SHA256).Hash.ToLower()
$apkName = "rain-qa-func-$vc.apk"
Write-Host "SHA256: $sha"

# Generate latest.json (matches Go UpdateService.UpdateInfo struct)
$json = @{
    versionCode  = $vc
    versionName  = $vn
    apkUrl       = "https://itsnot.fun/rain-qa-func/$apkName"
    sha256       = $sha
    releaseNotes = $Notes
} | ConvertTo-Json -Compress
$jsonPath = Join-Path $ROOT "build/latest.json"
[System.IO.File]::WriteAllText($jsonPath, $json, [System.Text.UTF8Encoding]::new($false))

# scp upload APK + latest.json
scp $APK "$REMOTE_DIR/$apkName"
scp $jsonPath "$REMOTE_DIR/latest.json"
Write-Host ""
Write-Host "Published: https://itsnot.fun/rain-qa-func/latest.json (versionCode=$vc)" -ForegroundColor Green
Write-Host "  APK: https://itsnot.fun/rain-qa-func/$apkName"
