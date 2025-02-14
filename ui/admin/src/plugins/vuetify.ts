// Styles
// import "@mdi/font/css/materialdesignicons.css";
import "vuetify/styles";
import { VNumberInput } from "vuetify/labs/VNumberInput";

// Vuetify
import { createVuetify } from "vuetify";

export default createVuetify({
  theme: {
    defaultTheme: "dark",
  },
  components: {
    VNumberInput,
  },
});
