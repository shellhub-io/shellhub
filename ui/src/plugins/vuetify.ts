// Styles
import "@mdi/font/css/materialdesignicons.css";
import "@fortawesome/fontawesome-free/css/all.css";
import "vuetify/styles";
import { VFileUpload, VFileUploadItem } from "vuetify/labs/VFileUpload";
import "../styles/_variables.scss";
import { aliases as faAliases, fa } from "vuetify/iconsets/fa";

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
    terminal: "#0f1526",
    "v-theme-background": "#FFFFFF",
    "v-theme-surface": "#FFFFFF",
    "v-theme-card": "#FFFFFF",
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
    terminal: "#0f1526",
    "v-theme-background": "#1E2127",
    "v-theme-surface": "#1E2127",
    "v-theme-card": "#22252B",
  },
};

export default createVuetify({
  components: {
    VFileUpload,
    VFileUploadItem,
  },
  theme: {
    defaultTheme: "dark",
    themes: {
      dark,
      light,
    },
  },
  defaults: {
    global: {

    },
    VTextField: {
      variant: "outlined",
      density: "comfortable",
    },
    VSelect: {
      variant: "outlined",
      density: "comfortable",
    },
    VTextarea: {
      variant: "outlined",
    },
  },
  icons: {
    faAliases,
    sets: {
      fa,
    },
  },
});
