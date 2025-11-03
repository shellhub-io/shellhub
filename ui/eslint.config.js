import { defineConfig, globalIgnores } from "eslint/config";
import js from "@eslint/js";
import typescriptEslint from "typescript-eslint";
import importPlugin from "eslint-plugin-import";
import jestPlugin from "eslint-plugin-jest";
import vuePlugin from "eslint-plugin-vue";
import prettierPlugin from "eslint-plugin-prettier";
import vueParser from "vue-eslint-parser";
import typeScriptParser from "@typescript-eslint/parser";
import globals from "globals";

export default defineConfig([
  globalIgnores(["eslint.config.js", "dist/", "node_modules/", "coverage/", "src/api/client/*.ts"]),

  { files: ["*.js", "*.cjs", "*.mjs", "*.ts", "*.vue"], },

  js.configs.recommended,
  typescriptEslint.configs.eslintRecommended,
  typescriptEslint.configs.recommended,
  typescriptEslint.configs.recommendedTypeChecked,

  {
    languageOptions: {
      parser: vueParser,
      parserOptions: {
        parser: typeScriptParser,
        project: "./tsconfig.json",
        extraFileExtensions: [".vue"],
        sourceType: "module",
        ecmaVersion: 2020,
        projectService: true,
      },
      globals: {
        ...globals.node,
        ...globals.jest,
      },
    },

    plugins: {
      import: importPlugin,
      vue: vuePlugin,
      prettier: prettierPlugin,
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

      "import/no-unresolved": "off",
      "import/no-extraneous-dependencies": ["error", { peerDependencies: true }],
      "import/no-cycle": ["off", { ignoreExternal: true }],
      "import/extensions": "off",
      "import/no-useless-path-segments": "off",

      "vue/max-len": ["error", { code: 140, template: 140 }],
      "vue/multi-word-component-names": "off",

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
      "@typescript-eslint/prefer-readonly-parameter-types": "off",
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

  // Tests override
  {
    files: ["**/*.spec.ts", "**/__tests__/*.{j,t}s?(x)", "**/tests/unit/**/*.spec.{j,t}s?(x)"],
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
      "@typescript-eslint/no-unsafe-call": "off",
      "@typescript-eslint/no-unsafe-member-access": "off",
      "@typescript-eslint/no-unsafe-assignment": "off",
      "jest/unbound-method": "error",
      "jest/no-focused-tests": "error",
    },
  },
]);
