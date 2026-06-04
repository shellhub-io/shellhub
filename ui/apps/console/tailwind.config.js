import preset from "@shellhub/design-system/tailwind.preset";

/** @type {import('tailwindcss').Config} */
export default {
  presets: [preset],
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
    "../../packages/design-system/**/*.{ts,tsx}",
  ],
  plugins: [],
};
