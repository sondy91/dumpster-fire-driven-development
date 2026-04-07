const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const PROJECT_ROOT = path.join(__dirname, '..');
const NODE_MODULES_DIR = path.join(PROJECT_ROOT, 'node_modules');
const APP_DIR = path.join(PROJECT_ROOT, 'app');
const OUTPUT_FILE = path.join(PROJECT_ROOT, 'scc-aggregate-results.txt');
const SCC_BIN = '/root/go/bin/scc';

console.log("🔥 Initializing SCC Batch Aggregator...");

// Initialize the output file
fs.writeFileSync(OUTPUT_FILE, "DUMPSTER FIRE DRIVEN DEVELOPMENT\n================================\nSCC AGGREGATE RESULTS\n\n");

let totalFiles = 0;
let totalLines = 0;
let totalBlanks = 0;
let totalComments = 0;
let totalCode = 0;

// 1. Analyze the core app and root files
console.log("Analyzing /app and root files...");
const appResult = execSync(`${SCC_BIN} ${APP_DIR} ${path.join(PROJECT_ROOT, 'PLAN.md')} ${path.join(PROJECT_ROOT, 'README.md')} ${path.join(PROJECT_ROOT, 'scripts')}`).toString();
fs.appendFileSync(OUTPUT_FILE, "--- FRONTEND & CONFIG ---\n");
fs.appendFileSync(OUTPUT_FILE, appResult + "\n");

// Parse JSON for the base files to add to our running total
const appJson = execSync(`${SCC_BIN} -f json ${APP_DIR} ${path.join(PROJECT_ROOT, 'PLAN.md')} ${path.join(PROJECT_ROOT, 'README.md')} ${path.join(PROJECT_ROOT, 'scripts')}`).toString();
JSON.parse(appJson).forEach(lang => {
    totalFiles += lang.Count;
    totalLines += lang.Lines;
    totalBlanks += lang.Blank;
    totalComments += lang.Comment;
    totalCode += lang.Code;
});

// 2. Iterate through node_modules carefully to prevent OOM
console.log("Iterating through massive node_modules structure...");
fs.appendFileSync(OUTPUT_FILE, "--- ENTERPRISE DEPENDENCIES (node_modules) ---\n");

const packages = fs.readdirSync(NODE_MODULES_DIR).filter(p => fs.statSync(path.join(NODE_MODULES_DIR, p)).isDirectory());

// To save the user's hard drive from a 500-page text file, we will aggregate the JSON data 
// programmatically, and only print a summarized output to the text file.
for (let i = 0; i < packages.length; i++) {
    const pkgPath = path.join(NODE_MODULES_DIR, packages[i]);
    const jsonOut = execSync(`${SCC_BIN} -f json ${pkgPath}`).toString();
    const data = JSON.parse(jsonOut);
    
    data.forEach(lang => {
        totalFiles += lang.Count;
        totalLines += lang.Lines;
        totalBlanks += lang.Blank;
        totalComments += lang.Comment;
        totalCode += lang.Code;
    });

    if ((i + 1) % 50 === 0) {
        console.log(`Processed ${i + 1}/${packages.length} packages...`);
    }
}

// 3. Write Final Summarized Block
const summary = `
================================
FINAL AGGREGATED SCOREBOARD
================================
Total Files:      ${totalFiles.toLocaleString()}
Total Lines:      ${totalLines.toLocaleString()}
Total Blanks:     ${totalBlanks.toLocaleString()}
Total Comments:   ${totalComments.toLocaleString()}
Total Code:       ${totalCode.toLocaleString()}

* Note: Standard SCC execution ran out of memory (OOM Exit Code 137). 
  This aggregation was achieved via iterative chunking. 
  You are officially a ${Math.floor(totalLines / 1000000)}-X Developer.
`;

fs.appendFileSync(OUTPUT_FILE, summary);

console.log("✅ Aggregation complete! Check scc-aggregate-results.txt");
console.log(summary);
