import js from "@eslint/js";
import { defineConfig, globalIgnores } from "eslint/config";
import tseslint from "typescript-eslint";
import reactHooks from "eslint-plugin-react-hooks";
import reactRefresh from "eslint-plugin-react-refresh";
import unusedImports from "eslint-plugin-unused-imports";
import stylisticPlugin from "@stylistic/eslint-plugin";
import vitestPlugin from "@vitest/eslint-plugin";
import globals from "globals";
import jsxA11y from "eslint-plugin-jsx-a11y";
import reactPlugin from "eslint-plugin-react";
import jsdocPlugin from "eslint-plugin-jsdoc";

export default defineConfig([
  globalIgnores(["**/dist", "**/node_modules", "**/.astro", "**/.vite", "**/src/client"]),
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

  jsxA11y.flatConfigs.recommended,
  {
    rules: {
      "jsx-a11y/control-has-associated-label": ["error", {
        ignoreElements: ["audio", "canvas", "embed", "input", "textarea", "tr", "video"],
        ignoreRoles: ["grid", "listbox", "menu", "menubar", "radiogroup", "row", "tablist", "toolbar", "tree", "treegrid"],
      }],

      "jsx-a11y/interactive-supports-focus": "warn",

      "jsx-a11y/no-static-element-interactions": "warn",

      "jsx-a11y/click-events-have-key-events": "warn",
    },
  },

  {
    plugins: { react: reactPlugin },
    settings: { react: { version: "detect" } },
    rules: {
      "react/button-has-type": "error",

      "react/no-danger": "warn",

      "react/jsx-no-target-blank": "error",

      "react/self-closing-comp": "warn",
    },
  },

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

      "@stylistic/indent": "off",
      "@stylistic/indent-binary-ops": "off",
      "@stylistic/jsx-indent-props": "off",
      "@stylistic/operator-linebreak": "off",
      "@stylistic/jsx-wrap-multilines": "off",
      "@stylistic/jsx-curly-newline": "off",
    },
  },

  {
    files: ["packages/design-system/**/*.{ts,tsx}"],
    rules: {
      "react-refresh/only-export-components": "off",
    },
  },

  {
    files: ["packages/design-system/**/*.{ts,tsx}"],
    ignores: ["**/__tests__/**", "**/*.test.ts", "**/*.test.tsx"],
    plugins: { jsdoc: jsdocPlugin },
    rules: {
      "jsdoc/require-jsdoc": ["error", {
        publicOnly: true,
        require: {
          ArrowFunctionExpression: true,
          ClassDeclaration: true,
          FunctionDeclaration: true,
          FunctionExpression: true,
          MethodDefinition: true,
        },
        contexts: [
          "ExportNamedDeclaration > TSInterfaceDeclaration",
          "ExportNamedDeclaration > TSTypeAliasDeclaration",
          "ExportNamedDeclaration > VariableDeclaration",
        ],
      }],
      "jsdoc/require-description": ["error", { contexts: ["any"] }],
    },
  },

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
