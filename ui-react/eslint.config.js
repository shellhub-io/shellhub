import js from "@eslint/js";
import { defineConfig, globalIgnores } from "eslint/config";
import tseslint from "typescript-eslint";
import reactHooks from "eslint-plugin-react-hooks";
import reactRefresh from "eslint-plugin-react-refresh";
import unusedImports from "eslint-plugin-unused-imports";
import stylisticPlugin from "@stylistic/eslint-plugin";
import vitestPlugin from "@vitest/eslint-plugin";
import globals from "globals";

export default defineConfig([
  globalIgnores(["**/dist", "**/node_modules", "**/.astro", "**/.vite", "packages/design-system/**", "**/src/client"]),
  js.configs.recommended,
  {
    extends: [...tseslint.configs.recommendedTypeChecked],
    languageOptions: {
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },
  stylisticPlugin.configs.recommended,
  {
    plugins: {
      "react-hooks": reactHooks,
      "react-refresh": reactRefresh,
      "unused-imports": unusedImports,
      "@stylistic": stylisticPlugin,
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
      "react-refresh/only-export-components": ["warn", { allowConstantExport: true }],

      "no-var": "error",
      "no-console": process.env.NODE_ENV === "production" ? "warn" : "off",
      "no-debugger": process.env.NODE_ENV === "production" ? "warn" : "off",

      "@typescript-eslint/no-unused-vars": "off",
      "unused-imports/no-unused-imports": "error",
      "unused-imports/no-unused-vars": [
        "error",
        {
          argsIgnorePattern: "^_",
          varsIgnorePattern: "^_",
          caughtErrorsIgnorePattern: "^_",
        },
      ],

      "@stylistic/multiline-ternary": "off",
      "@stylistic/jsx-one-expression-per-line": "off",
      "@stylistic/quotes": ["error", "double", { avoidEscape: true }],
      "@stylistic/semi": ["error", "always"],
      "@stylistic/brace-style": ["error", "1tbs", { allowSingleLine: true }],
      "@stylistic/max-statements-per-line": "off",
      "@stylistic/member-delimiter-style": "off",
      "@stylistic/arrow-parens": ["error", "always"],
      "@stylistic/quote-props": ["error", "as-needed"],

      // Disabled — Prettier controls these and cannot be configured to match
      "@stylistic/indent": "off",
      "@stylistic/indent-binary-ops": "off",
      "@stylistic/jsx-indent-props": "off",
      "@stylistic/operator-linebreak": "off",
      "@stylistic/jsx-wrap-multilines": "off",
      "@stylistic/jsx-curly-newline": "off",
    },
  },

  // Disable type-checked rules for JS config files
  {
    files: ["**/*.js", "**/*.mjs", "**/*.cjs"],
    ...tseslint.configs.disableTypeChecked,
    languageOptions: {
      ...tseslint.configs.disableTypeChecked.languageOptions,
      globals: {
        ...globals.node,
      },
      parserOptions: {
        ...tseslint.configs.disableTypeChecked.languageOptions?.parserOptions,
        projectService: false,
      },
    },
  },

  // Test overrides
  {
    files: ["**/__tests__/**", "**/*.test.ts", "**/*.test.tsx"],
    plugins: {
      vitest: vitestPlugin,
    },
    rules: {
      "vitest/no-focused-tests": "error",
      "@typescript-eslint/unbound-method": "off",
      "@typescript-eslint/no-unsafe-call": "off",
      "@typescript-eslint/no-unsafe-member-access": "off",
      "@typescript-eslint/no-unsafe-assignment": "off",
      "@typescript-eslint/no-unsafe-argument": "off",
      "@typescript-eslint/require-await": "off",
    },
  },
]);
