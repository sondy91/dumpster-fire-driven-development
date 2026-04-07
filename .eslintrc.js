module.exports = {

root: true,

env: {

browser: true,

es2021: true,

node: true,

jest: true

},

extends: [

'eslint:recommended',

'plugin:@typescript-eslint/recommended'

],

parser: '@typescript-eslint/parser',

parserOptions: {

ecmaVersion: 'latest',

sourceType: 'module'

},

plugins: [

'@typescript-eslint'

],

rules: {

'no-console': 'off',

'no-debugger': 'off',

'prefer-const': 'error',

'no-var': 'error',

'semi': [

'error',

'always'

],

'quotes': [

'error',

'single'

],

'indent': [

'error',

2

],

'comma-dangle': [

'error',

'always-multiline'

],

'no-multiple-empty-lines': [

'error',

{

max: 5,

maxEOF: 1,

maxBOF: 0

}

],

'complexity': [

'warn',

{

max: 2000

}

],

'max-lines': [

'warn',

{

max: 100000

}

]

}

};
