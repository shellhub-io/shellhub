import { createApp } from "vue";
import vuetify from "@ui/plugins/vuetify";
import App from "./App.vue";
import { loadFonts } from "./plugins/webfontloader";
import { key, store } from "./store";
import router from "./router";
import SnackbarComponent from "./components/Snackbar/Snackbar.vue";

const app = createApp(App);

loadFonts();

app.use(vuetify);
app.use(router);
app.use(store, key);
app.component("SnackbarComponent", SnackbarComponent);
app.mount("#app");
