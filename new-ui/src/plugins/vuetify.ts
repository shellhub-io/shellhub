// Styles
import "@mdi/font/css/materialdesignicons.css";
import "vuetify/styles";
import "../styles/_variables.scss";

// Vuetify
import { createVuetify } from "vuetify";

const light = {
  dark: false,
  colors: {
    primary: "#667acc",
    secondary: "#FFFFFF",
    background: "#F5F5F5",
    tabs: "#F5F5F5",
    foreground: "#F5F5F5",
    paymentForm: "#F5F5F5",
    "v-theme-background":" #FFFFFF",
    "v-theme-surface":  "#FFFFFF",
  },
};

const dark = {
  dark: true,
  colors: {
    primary: "#667acc",
    secondary: "#1E2127",
    background: "#18191B",
    tabs: "#1E1E1E",
    foreground: "#1E1E1E",
    paymentForm: "#E0E0E0",
    "v-theme-background":" #1E2127",
    "v-theme-surface":  "#1E2127",
  },
};

export default createVuetify({
  theme: {
    defaultTheme: "dark",
    themes: {
      dark,
      light,
    },
  },
});
