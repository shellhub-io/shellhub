/** @type {import('tailwindcss').Config} */
export default {
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: "#667ACC",
          50: "#667ACC0d",
          100: "#667ACC1a",
          200: "#667ACC33",
          300: "#667ACC4d",
          400: "#667ACC80",
          500: "#667ACC",
          600: "#5468b3",
          700: "#4a5c9e",
        },
        // Semantic tokens are driven by CSS variables (see design-system/css/base.css)
        // so the whole UI can flip between dark (default) and light themes.
        background: "rgb(var(--c-background) / <alpha-value>)",
        surface: "rgb(var(--c-surface) / <alpha-value>)",
        card: "rgb(var(--c-card) / <alpha-value>)",
        border: "rgb(var(--c-border) / <alpha-value>)",
        "border-light": "rgb(var(--c-border-light) / <alpha-value>)",
        "text-primary": "rgb(var(--c-text-primary) / <alpha-value>)",
        "text-secondary": "rgb(var(--c-text-secondary) / <alpha-value>)",
        "text-muted": "rgb(var(--c-text-muted) / <alpha-value>)",
        accent: {
          green: "#82a568",
          red: "#D8737B",
          yellow: "#bf8c5d",
          blue: "#56a2e1",
          cyan: "#4e9aa3",
        },
        // Hover overlays flip from white-on-dark to black-on-light via --c-hover.
        "hover-subtle": "rgb(var(--c-hover) / 0.03)",
        "hover-medium": "rgb(var(--c-hover) / 0.05)",
        "hover-strong": "rgb(var(--c-hover) / 0.08)",
      },
      fontFamily: {
        sans: ['"IBM Plex Sans"', "system-ui", "sans-serif"],
        mono: ['"IBM Plex Mono"', "monospace"],
      },
      fontSize: {
        "3xs": ["0.5625rem", { lineHeight: "0.75rem" }],
        "2xs": ["0.625rem", { lineHeight: "0.875rem" }],
      },
      letterSpacing: {
        compact: "0.10em",
        label: "0.15em",
        wide: "0.20em",
      },
      opacity: {
        faint: "0.20",
        soft: "0.30",
        dim: "0.40",
      },
      backgroundImage: {
        "gradient-radial": "radial-gradient(ellipse at center, var(--tw-gradient-stops))",
      },
      animation: {
        "fade-in": "fadeIn 0.4s ease-out",
        "slide-up": "slideUp 0.35s ease-out",
        "slide-down": "slideDown 0.2s ease-out",
        "pulse-subtle": "pulseSubtle 2s ease-in-out infinite",
        shake: "shake 0.4s ease-in-out",
      },
      keyframes: {
        fadeIn: {
          "0%": { opacity: "0" },
          "100%": { opacity: "1" },
        },
        slideUp: {
          "0%": { opacity: "0", transform: "translateY(8px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
        slideDown: {
          "0%": { opacity: "0", transform: "translateY(-4px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
        pulseSubtle: {
          "0%, 100%": { opacity: "1" },
          "50%": { opacity: "0.5" },
        },
        shake: {
          "0%, 100%": { transform: "translateX(0)" },
          "20%, 60%": { transform: "translateX(-4px)" },
          "40%, 80%": { transform: "translateX(4px)" },
        },
      },
    },
  },
};
