import { createApp, type Plugin } from "vue";
import { createPinia } from "pinia";
import vuetify from "@/plugins/vuetify";
import App from "./App.vue";
import loadFonts from "@/plugins/webfontloader";
import router from "./router";
import { SnackbarPlugin } from "@/plugins/snackbar";

const pinia = createPinia();
const app = createApp(App);

loadFonts();
app.use(vuetify as Plugin);
app.use(router);
app.use(pinia);
app.use(SnackbarPlugin);
app.mount("#app");
