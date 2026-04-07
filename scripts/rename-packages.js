const fs = require('fs');
const path = require('path');

const PROJECT_ROOT = path.join(__dirname, '..');
const NODE_MODULES_DIR = path.join(PROJECT_ROOT, 'node_modules');

const prefixes = ['react', 'vue', 'angular', 'webpack', 'babel', 'jest', 'eslint', 'rxjs', 'redux', 'graphql', 'is', 'left-pad', 'aws', 'firebase', 'tailwind', 'bootstrap', 'enterprise', 'synergy', 'next', 'nuxt', 'nest', 'svelte', 'express', 'koa', 'fastify', 'prisma', 'apollo', 'mongoose'];
const suffixes = ['vibe', 'wrapper', 'loader', 'plugin', 'blockchain', 'enterprise', 'spaghetti', 'generator', 'operator', 'overkill', 'schema', 'destroyer', 'manager', 'factory', 'provider', 'strategy', 'polyfill', 'polyfill-v2', 'adapter', 'bridge', 'facade', 'proxy', 'observer', 'memento'];

const packages = fs.readdirSync(NODE_MODULES_DIR).filter(p => fs.statSync(path.join(NODE_MODULES_DIR, p)).isDirectory());

console.log("🚀 Renaming packages to maximum buzzword compliance...");

let renamedCount = 0;

packages.forEach((pkg) => {
    if (pkg.startsWith('enterprise-validator-')) {
        const id = pkg.replace('enterprise-validator-', '');
        const prefix = prefixes[Math.floor(Math.random() * prefixes.length)];
        const suffix1 = suffixes[Math.floor(Math.random() * suffixes.length)];
        const suffix2 = suffixes[Math.floor(Math.random() * suffixes.length)];
        const newName = `${prefix}-${suffix1}-${suffix2}-${id}`;
        
        const oldPath = path.join(NODE_MODULES_DIR, pkg);
        const newPath = path.join(NODE_MODULES_DIR, newName);
        
        fs.renameSync(oldPath, newPath);
        
        const indexPath = path.join(newPath, 'index.js');
        if (fs.existsSync(indexPath)) {
            let content = fs.readFileSync(indexPath, 'utf8');
            // Update class name to something absurdly camelCased
            const className = newName.split('-').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join('');
            // The original generated class name was EnterpriseValidator_<id>
            content = content.replace(new RegExp(`EnterpriseValidator_${id}`, 'g'), className);
            fs.writeFileSync(indexPath, content);
        }
        
        renamedCount++;
    }
});

console.log(`✅ Successfully synergized ${renamedCount} package names!`);
