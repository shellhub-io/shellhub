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

  // Accessibility — catches missing aria-labels, bad roles, inaccessible interactive elements
  jsxA11y.flatConfigs.recommended,
  {
    rules: {
      // Icon-only interactive elements must have an accessible label.
      // Catches close buttons, icon buttons, etc. that render no visible text.
      "jsx-a11y/control-has-associated-label": ["error", {
        ignoreElements: ["audio", "canvas", "embed", "input", "textarea", "tr", "video"],
        ignoreRoles: ["grid", "listbox", "menu", "menubar", "radiogroup", "row", "tablist", "toolbar", "tree", "treegrid"],
      }],

      // Interactive elements (role=button, etc.) must be keyboard-focusable.
      // Catches <div onClick> that omit tabIndex and keyboard handlers.
      "jsx-a11y/interactive-supports-focus": "warn",

      // Prevents attaching mouse/pointer handlers to static, non-interactive elements
      // without a corresponding keyboard handler.
      "jsx-a11y/no-static-element-interactions": "warn",

      // Any element with an onClick must also handle Enter/Space via onKeyDown/onKeyUp/onKeyPress.
      "jsx-a11y/click-events-have-key-events": "warn",
    },
  },

  // React — button types, dangerous HTML, safe anchor targets
  {
    plugins: { react: reactPlugin },
    settings: { react: { version: "detect" } },
    rules: {
      // Same goal as jsx-a11y/button-has-type but from the React side.
      // Together they cover both JSX and DOM-level checks.
      "react/button-has-type": "error",

      // Prevents dangerouslySetInnerHTML from being introduced silently.
      "react/no-danger": "warn",

      // <a target="_blank"> without rel="noreferrer" is a security/privacy issue.
      "react/jsx-no-target-blank": "error",

      // Enforces self-closing tags on components/elements with no children.
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
