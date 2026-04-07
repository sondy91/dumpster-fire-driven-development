const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

/**
 * Enterprise Commit Spammer v2.0
 * Now with --dry-run and synergistic commit messaging.
 */

const PROJECT_ROOT = path.join(__dirname, '..');
const COMMITS_DIR = path.join(PROJECT_ROOT, 'commits-hell');
const TARGET_COMMITS = 1000;

// All the classics from the Bloat Engine and then some
const TERRIBLE_MESSAGES = [
    "Update", "fix", "wip", "asdfghjkl", "I hate this codebase",
    "Merge branch 'master' of https://github.com/sondy91/dumpster into master",
    "Removed a console.log", "final version", "final version for real",
    "final version 2 DO NOT TOUCH", "...", "Revert \"Revert \"fix\"\"",
    "pls work", "more code = more assets", "Enterprise architecture scaling",
    "Add synergized dependency", "Typo", "Linting", "vibes", "vibe coding",
    "Ship it", "LFG", "10x dev moment", "🚀", "📈", "Grindset", "OOM",
    "Refactor", "Cleanup", "Fixed a bug", "Hotfix", "Tweak", "Minor changes"
];

const alphabet = 'abcdefghijklmnopqrstuvwxyz'.split('');
const isDryRun = process.argv.includes('--dry-run');

function getRandomAlphabetMessage() {
    const len = Math.floor(Math.random() * 3) + 1;
    let msg = '';
    for (let i = 0; i < len; i++) {
        msg += alphabet[Math.floor(Math.random() * alphabet.length)];
    }
    return msg;
}

function runGitCommand(command) {
    if (isDryRun) {
        console.log(`[DRY RUN] Executing: ${command}`);
        return;
    }
    try {
        execSync(command, { cwd: PROJECT_ROOT, stdio: 'ignore' });
    } catch (e) {
        // Robustness
    }
}

console.log("🚀 STARTING THE ATOMIC COMMIT SPAMMER...");
if (isDryRun) console.log("⚠️ DRY RUN MODE ENABLED - No files will be written, no commits made.");

if (!isDryRun && !fs.existsSync(COMMITS_DIR)) {
    fs.mkdirSync(COMMITS_DIR);
}

for (let i = 1; i <= TARGET_COMMITS; i++) {
    const msg = Math.random() > 0.5 
        ? TERRIBLE_MESSAGES[Math.floor(Math.random() * TERRIBLE_MESSAGES.length)]
        : getRandomAlphabetMessage();
        
    const filePath = path.join(COMMITS_DIR, `commit-${i}.txt`);
    const content = `Atomic Vibe #${i}\nTimestamp: ${Date.now()}\nMessage: ${msg}`;

    if (!isDryRun) {
        fs.writeFileSync(filePath, content);
    }

    runGitCommand(`git add ${path.relative(PROJECT_ROOT, filePath)}`);
    runGitCommand(`git commit -m "${msg}"`);
    
    if (i % 100 === 0) {
        console.log(`Processed ${i}/${TARGET_COMMITS} commits...`);
    }
}

console.log("🔥 ATOMIC COMMIT SPAMMER FINISHED 🔥");
