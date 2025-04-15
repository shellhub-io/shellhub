import { createApp } from "vue";
import { createPinia } from "pinia";
import vuetify from "./plugins/vuetify";
import App from "./App.vue";
import { loadFonts } from "./plugins/webfontloader";
import router from "./router";
import SnackbarComponent from "./components/Snackbar/Snackbar.vue";

const pinia = createPinia();
const app = createApp(App);

loadFonts();
app.use(vuetify);
app.use(router);
app.use(pinia);

app.component("SnackbarComponent", SnackbarComponent);
app.mount("#app");
