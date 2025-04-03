import { createApp } from "vue";
import vuetify from "@ui/plugins/vuetify";
import * as Globals from "@global/components/index";
import { createPinia } from "pinia";
import App from "./App.vue";
import { loadFonts } from "./plugins/webfontloader";
import { key, store } from "./store";
import router from "./router";
import SnackbarComponent from "./components/Snackbar/Snackbar.vue";

const pinia = createPinia();
const app = createApp(App);

loadFonts();

app.use(vuetify);
app.use(router);
app.use(store, key);
app.use(pinia);

Object.entries(Globals).forEach(([name, component]) => {
  app.component(name, component);
});

app.component("SnackbarComponent", SnackbarComponent);
app.mount("#app");
