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
        background: "#18191B",
        surface: "#1E2127",
        card: "#22252B",
        border: "#2C2F36",
        "border-light": "#383D47",
        "text-primary": "#E1E4EA",
        "text-secondary": "#8B8F99",
        "text-muted": "#81879C",
        accent: {
          green: "#82a568",
          red: "#D8737B",
          yellow: "#bf8c5d",
          blue: "#56a2e1",
          cyan: "#4e9aa3",
        },
        "hover-subtle": "rgba(255, 255, 255, 0.03)",
        "hover-medium": "rgba(255, 255, 255, 0.05)",
        "hover-strong": "rgba(255, 255, 255, 0.08)",
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
      },
    },
  },
};
