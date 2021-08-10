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
    indent: [2, 2],
    'no-cond-assign': ['error'],
    'no-constant-condition': ['error'],
    'no-empty-pattern': ['error'],
    'no-redeclare': ['error'],
    'no-delete-var': ['error'],
    'no-var': ['error'],
    'import/no-unresolved': 'off',
    'import/no-extraneous-dependencies': ['error', { peerDependencies: true }],
    'import/no-cycle': [0, { ignoreExternal: true }],
    'import/extensions': 'off',
    'spaced-comment': [2, 'always', {
      exceptions: ['////'],
      markers: ['/'],
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
