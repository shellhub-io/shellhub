module.exports = {
  root: true,
  env: {
    // this section will be used to determine which APIs are available to us
    // (i.e are we running in a browser environment or a node.js env)
    node: true,
    browser: true
  },
  parserOptions: {
    parser: "babel-eslint",
    // specifying a module sourcetype prevent eslint from marking import statements as errors
    sourceType: "module"
  },
  extends: [
    // use the recommended rule set for both plain javascript and vue
    "eslint:recommended",
    "plugin:vue/recommended"
  ],
  rules: {
    "semi": ["error", "always"],
    "quotes": ["error", "single", {"avoidEscape": true, "allowTemplateLiterals": true}],
    "indent": [2, 2],
    "no-duplicate-imports": ["error", { "includeExports": true }],
    "eqeqeq": ["error", "always"],
    "no-unused-vars": ["error"],
    "no-cond-assign": ["error"],
    "no-constant-condition": ["error"],
    "no-case-declarations": ["error"],
    "no-empty-pattern": ["error"],
    "no-redeclare": ["error"],
    "no-delete-var": ["error"],
    "camelcase": ["error"],
    "brace-style": ["error"],
    "arrow-parens": ["error", "always"],
    "no-var": ["error"]
  }
};