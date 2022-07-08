/* eslint-env node */
require("@rushstack/eslint-patch/modern-module-resolution");

module.exports = {
  root: true,

  env: {
    node: true,
    jest: true,
  },

  extends: [
    "eslint:recommended",
    "plugin:vue/vue3-essential",
    "@vue/airbnb",
    "@vue/eslint-config-typescript/recommended",
  ],

  parserOptions: {
    ecmaVersion: 2020,
  },
  rules: {
    quotes: [2, "double", "avoid-escape"],
    "no-console": process.env.NODE_ENV === "production" ? "warn" : "off",
    "no-debugger": process.env.NODE_ENV === "production" ? "warn" : "off",
    "global-require": 0,
    indent: [2],
    "no-cond-assign": ["error"],
    "no-constant-condition": ["error"],
    "no-empty-pattern": ["error"],
    "no-redeclare": ["error"],
    "no-delete-var": ["error"],
    "no-var": ["error"],
    "import/no-unresolved": "off",
    "import/no-extraneous-dependencies": ["error", { peerDependencies: true }],
    "import/no-cycle": [0, { ignoreExternal: true }],
    "import/extensions": "off",
    "max-len": ["error", { code: 140 }],
    "spaced-comment": [
      2,
      "always",
      {
        exceptions: ["////"],
        markers: ["/"],
      },
    ],
    "import/no-useless-path-segments": [
      0,
      {
        noUselessIndex: true,
      },
    ],
    "vue/multi-word-component-names": [
      0,
      {
        ignores: [],
      },
    ],
    "no-shadow": [0, { hoist: "never" }],
    "no-confusing-arrow": [0, { allowParens: true, onlyOneSimpleParam: false }],
    "object-curly-newline": [0, "always"],
    "no-plusplus": 0,
  },

  overrides: [
    {
      files: ["**/__tests__/*.{j,t}s?(x)", "**/tests/unit/**/*.spec.{j,t}s?(x)"],
      env: {
        jest: true,
      },
    },
    {
      files: ["**/__tests__/*.{j,t}s?(x)", "**/tests/unit/**/*.spec.{j,t}s?(x)"],
      env: {
        jest: true,
      },
    },
    {
      files: ["**/__tests__/*.{j,t}s?(x)", "**/tests/unit/**/*.spec.{j,t}s?(x)"],
      env: {
        jest: true,
      },
    },
  ],
};
