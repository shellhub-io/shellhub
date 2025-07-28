// eslint.config.mjs
import { defineConfig } from "eslint/config";
import js from "@eslint/js";
import typescriptEslint from "typescript-eslint";
import importPlugin from "eslint-plugin-import";
import jestPlugin from "eslint-plugin-jest";
import vuePlugin from "eslint-plugin-vue";
import prettier from "eslint-config-prettier";
import globals from "globals";

export default defineConfig([
  // Prevent ESLint from linting itself
  {
    ignores: ["eslint.config.mjs"],
  },

  // Base recommendations
  js.configs.recommended,
  typescriptEslint.configs.eslintRecommended,
  typescriptEslint.configs.recommended,
  typescriptEslint.configs.recommendedTypeChecked,
  prettier,

  // Main config
  {
    languageOptions: {
      parserOptions: {
        project: "./tsconfig.json",
        sourceType: "module",
        ecmaVersion: 2020,
      },
      globals: {
        ...globals.node,
        ...globals.jest,
      },
    },

    plugins: {
      import: importPlugin,
      vue: vuePlugin,
    },

    rules: {
      quotes: ["error", "double", { avoidEscape: true }],
      "no-console": process.env.NODE_ENV === "production" ? "warn" : "off",
      "no-debugger": process.env.NODE_ENV === "production" ? "warn" : "off",
      "global-require": "off",
      indent: ["error", 2],
      "no-cond-assign": ["error"],
      "no-constant-condition": ["error"],
      "no-empty-pattern": ["error"],
      "no-redeclare": ["error"],
      "no-delete-var": ["error"],
      "no-var": ["error"],
      "no-plusplus": "off",
      "no-shadow": "off",
      "no-confusing-arrow": "off",
      "object-curly-newline": "off",

      // Import plugin rules
      "import/no-unresolved": "off",
      "import/no-extraneous-dependencies": ["error", { peerDependencies: true }],
      "import/no-cycle": ["off", { ignoreExternal: true }],
      "import/extensions": "off",
      "import/no-useless-path-segments": "off",

      // Vue plugin rules
      "vue/max-len": ["error", { code: 140, template: 140 }],
      "vue/multi-word-component-names": "off",

      // Comments
      "spaced-comment": [
        "error",
        "always",
        {
          exceptions: ["////"],
          markers: ["/"],
        },
      ],
      
      "@typescript-eslint/no-shadow": "off",
      "@typescript-eslint/no-use-before-define": "off",
      "@typescript-eslint/require-await": "off",
      "@typescript-eslint/prefer-readonly-parameter-types": "off",
      "@typescript-eslint/no-throw-literal": "off",
      "@typescript-eslint/no-unused-vars": [
        "error",
        {
          argsIgnorePattern: "^_",
          varsIgnorePattern: "^_",
          caughtErrorsIgnorePattern: "^_",
        },
      ],
    },
  },

  // Jest tests override
  {
    files: ["**/*.test.ts", "**/__tests__/*.{j,t}s?(x)", "**/tests/unit/**/*.spec.{j,t}s?(x)"],
    plugins: {
      jest: jestPlugin,
    },
    languageOptions: {
      globals: {
        ...globals.jest,
      },
    },
    rules: {
      "@typescript-eslint/unbound-method": "off",
      "jest/unbound-method": "error",
      "jest/no-focused-tests": "error",
    },
  },
]);
