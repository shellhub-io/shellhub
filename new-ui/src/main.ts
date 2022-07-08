import { createApp } from "vue";
import App from "./App.vue";
import vuetify from "./plugins/vuetify";
import { key, store } from "./store";
import router from "./router";

import { loadFonts } from "./plugins/webfontloader";

import SnackbarComponent from "./components/Snackbar/Snackbar.vue";

loadFonts();

createApp(App)
  .use(vuetify)
  .use(router)
  .use(store, key)
  .component("SnackbarComponent", SnackbarComponent)
  .mount("#app");
