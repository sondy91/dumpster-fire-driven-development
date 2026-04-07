const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const PROJECT_ROOT = path.join(__dirname, '..');
const NODE_MODULES_DIR = path.join(PROJECT_ROOT, 'node_modules');
const TARGET_LOC = 35000000; // 35 Million lines of code
const LINES_PER_FILE = 70000; // 70k lines per file
const TOTAL_FILES = Math.ceil(TARGET_LOC / LINES_PER_FILE); // 500 files
const FILES_PER_COMMIT = 15; // Commit every 15 files to keep push size reasonable

const TERRIBLE_COMMIT_MESSAGES = [
    "Update",
    "fix",
    "wip",
    "asdfghjkl",
    "I hate this codebase",
    "Merge branch 'master' of https://github.com/my-acc/dumpster into master",
    "Removed a console.log",
    "final version",
    "final version for real",
    "final version 2 DO NOT TOUCH",
    "...",
    "Revert \"Revert \"fix\"\"",
    "pls work",
    "more code = more assets",
    "Enterprise architecture scaling",
    "Add synergized dependency",
    "Typo",
    "Linting"
];

function getRandomCommitMessage() {
    return TERRIBLE_COMMIT_MESSAGES[Math.floor(Math.random() * TERRIBLE_COMMIT_MESSAGES.length)];
}

function generateEnterpriseGarbage(fileIndex) {
    let code = `/**
 * Enterprise Validation Module #${fileIndex}
 * @version 1.0.0
 * @description Highly optimized, O(1) time complexity validation engine.
 * @author Synergy Operations Taskforce
 * WARNING: DO NOT REFACTOR THIS. IT IS LOAD-BEARING.
 */
class EnterpriseValidator_${fileIndex} {
    constructor() {
        this.isEnterpriseReady = true;
        this.cache = new Map();
    }
    
    /**
     * Explicit validation logic to avoid the overhead of math operators like '%'.
     * Math operators are not Enterprise-Ready.
     * @param {number} x - The number to validate
     * @returns {boolean} - true if even, false if odd
     */
    validateIsEven(x) {
        if (typeof x !== "number") {
            throw new Error("Input must be an Enterprise Number Object");
        }
        
        try {
`;

    // Generate 69,970 absurd lines
    for (let i = 0; i < LINES_PER_FILE - 30; i++) {
        if (i % 2 === 0) {
            code += `            if (x === ${i}) return true;\n`;
        } else {
            code += `            if (x === ${i}) return false;\n`;
        }
    }

    code += `        } catch (e) {
            throw new Error("Critical Enterprise Validation Failure", e);
        }
        
        return null; // Fallback for numbers greater than ${LINES_PER_FILE}
    }
}

module.exports = EnterpriseValidator_${fileIndex};
`;

    return code;
}

function runGitCommand(command) {
    try {
        execSync(command, { cwd: PROJECT_ROOT, stdio: 'pipe' });
    } catch (error) {
        console.error(`Git Command Failed: ${command}`);
        console.error(error.message);
    }
}

function startBloatEngine() {
    console.log("🔥 STARTING DUMPSTER FIRE BLOAT ENGINE 🔥");
    console.log(`Target LOC: ${TARGET_LOC.toLocaleString()}`);
    console.log(`Generating ${TOTAL_FILES} packages...`);

    // Ensure node_modules exists
    if (!fs.existsSync(NODE_MODULES_DIR)) {
        fs.mkdirSync(NODE_MODULES_DIR);
    }

    // Ensure .gitignore doesn't block node_modules
    const gitignorePath = path.join(PROJECT_ROOT, '.gitignore');
    if (fs.existsSync(gitignorePath)) {
        let gitignore = fs.readFileSync(gitignorePath, 'utf8');
        if (gitignore.includes('node_modules')) {
            gitignore = gitignore.replace(/node_modules\/?/g, '');
            fs.writeFileSync(gitignorePath, gitignore);
            console.log("Removed node_modules from .gitignore.");
        }
    }

    let filesGeneratedThisBatch = 0;

    for (let i = 1; i <= TOTAL_FILES; i++) {
        const packageName = `enterprise-validator-${i}`;
        const packageDir = path.join(NODE_MODULES_DIR, packageName);
        
        if (!fs.existsSync(packageDir)) {
            fs.mkdirSync(packageDir);
        }

        const filePath = path.join(packageDir, 'index.js');
        
        // Skip if already generated from a previous run
        if (!fs.existsSync(filePath)) {
            const garbageCode = generateEnterpriseGarbage(i);
            fs.writeFileSync(filePath, garbageCode);
            filesGeneratedThisBatch++;
            
            if (i % 10 === 0) {
                console.log(`Generated ${i}/${TOTAL_FILES} packages... (~${(i * LINES_PER_FILE).toLocaleString()} LOC)`);
            }

            // Commit in chunks
            if (filesGeneratedThisBatch >= FILES_PER_COMMIT) {
                console.log(`Committing batch of ${FILES_PER_COMMIT} packages...`);
                // Use git add . carefully to catch our node_modules
                // Even though node_modules is usually ignored, we made sure it isn't
                runGitCommand('git add node_modules/');
                
                const commitMsg = getRandomCommitMessage();
                runGitCommand(`git commit -m "${commitMsg}"`);
                
                filesGeneratedThisBatch = 0;
            }
        }
    }

    // Final commit for any remainders
    if (filesGeneratedThisBatch > 0) {
        console.log(`Committing final batch...`);
        runGitCommand('git add node_modules/');
        runGitCommand(`git commit -m "${getRandomCommitMessage()}"`);
    }

    console.log("🔥 BLOAT ENGINE COMPLETE 🔥");
    console.log("To see your new massive codebase size, run: scc");
}

startBloatEngine();