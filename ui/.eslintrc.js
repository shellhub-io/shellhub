module.exports = {
  root: true,
  env: {
    // this section will be used to determine which APIs are available to us
    // (i.e are we running in a browser environment or a node.js env)
    node: true,
    browser: true,
    jest: true,
  },
  parserOptions: {
    parser: 'babel-eslint',
    // specifying a module sourcetype prevent eslint from marking import statements as errors
    sourceType: 'module',
  },
  extends: [
    // use the recommended rule set for both plain javascript and vue
    'eslint:recommended',
    'airbnb-base',
    'plugin:vue/recommended',
  ],
  plugins: ['import'],
  rules: {
    'global-require': 0,
    semi: ['error', 'always'],
    quotes: ['error', 'single', { avoidEscape: true, allowTemplateLiterals: true }],
    indent: [2, 2],
    'no-duplicate-imports': ['error', { includeExports: true }],
    eqeqeq: ['error', 'always'],
    'no-unused-vars': ['error'],
    'no-cond-assign': ['error'],
    'no-constant-condition': ['error'],
    'no-case-declarations': ['error'],
    'no-empty-pattern': ['error'],
    'no-redeclare': ['error'],
    'no-delete-var': ['error'],
    camelcase: ['error'],
    'brace-style': ['error'],
    'arrow-parens': ['error', 'always'],
    'no-var': ['error'],
    'import/no-unresolved': 'off',
    'import/no-extraneous-dependencies': ['error', { peerDependencies: true }],
    'import/no-cycle': [0, { ignoreExternal: true }],
    'import/extensions': ['error', 'ignorePackages', {
      js: 'never',
      vue: 'never',
    }],
  },
  settings: {
    'import/resolver': ['webpack', {
      node: {
        extensions: ['.js', '.vue'],
        moduleDirectory: ['node_modules', 'src/'],
      },
    },
    ],
  },
};
