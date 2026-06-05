import preset from "@shellhub/design-system/tailwind.preset";

export default {
  presets: [preset],
  content: [
    "./src/**/*.{astro,html,js,jsx,md,mdx,ts,tsx}",
    "../../packages/design-system/**/*.{ts,tsx}",
  ],
};
