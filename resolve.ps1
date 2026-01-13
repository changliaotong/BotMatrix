# Resolve UU conflicts (Keep Ours for non-BotWorker, Keep Theirs for BotWorker)
$uu_files = git status --short | Where-Object { $_ -match "^UU" } | ForEach-Object { $_.Substring(3).Trim() }
foreach ($file in $uu_files) {
    if ($file -like "src/BotWorker/*") {
        git checkout --theirs $file
    } else {
        git checkout --ours $file
    }
    git add $file
}

# Resolve DU conflicts (Keep Theirs for BotWorker, Keep Ours/Discard for others)
$du_files = git status --short | Where-Object { $_ -match "^DU" } | ForEach-Object { $_.Substring(3).Trim() }
foreach ($file in $du_files) {
    if ($file -like "src/BotWorker/*") {
        git checkout --theirs $file
        git add $file
    } else {
        git rm $file
    }
}

# Resolve UD conflicts (Keep Ours for non-BotWorker, Keep Theirs for BotWorker)
$ud_files = git status --short | Where-Object { $_ -match "^UD" } | ForEach-Object { $_.Substring(3).Trim() }
foreach ($file in $ud_files) {
    if ($file -like "src/BotWorker/*") {
        git rm $file
    } else {
        git checkout --ours $file
        git add $file
    }
}

# Final add all remaining changes
git add .

# Check if there are still unmerged files
$remaining = git status --short | Where-Object { $_ -match "^U" }
if ($remaining) {
    Write-Host "Remaining conflicts:"
    $remaining
} else {
    Write-Host "All conflicts resolved."
}
