import { createApp } from "vue";
import * as Sentry from "@sentry/vue";
import { BrowserTracing } from "@sentry/vue";
import VueGtag from "vue-gtag";
import { envVariables } from "./envVariables";
import vuetify from "./plugins/vuetify";
import { key, store } from "./store";
import { router } from "./router";
import App from "./App.vue";

// Local and Xterm Fonts
import { loadFonts } from "./plugins/webfontloader";
import "@fontsource/fira-code";
import "@fontsource/source-code-pro";
import "@fontsource/jetbrains-mono";
import "@fontsource/ubuntu-mono";
import "@fontsource/noto-mono";
import "@fontsource/inconsolata";
import "@fontsource/anonymous-pro";

import SnackbarComponent from "./components/Snackbar/Snackbar.vue";
import { SnackbarPlugin } from "./plugins/snackbar";

/* import the fontawesome core */

/* import font awesome icon component */

const app = createApp(App);

Sentry.init({
  app,
  dsn: envVariables.sentryDsn || "",
  integrations: [
    new BrowserTracing({
      routingInstrumentation: Sentry.vueRouterInstrumentation(router),
    }),
  ],
  tracesSampleRate: 1.0,
  hooks: ["activate", "create", "destroy", "mount", "update"],
  timeout: 500,
  release: envVariables.version || "latest",
});
Sentry.setTag("project", "shellhub-ui");

loadFonts();

app.use(vuetify);
app.use(router);
app.use(store, key);
app.use(VueGtag, {
  config: { id: envVariables.googleAnalyticsID || "" },
});
app.use(SnackbarPlugin);
app.component("SnackbarComponent", SnackbarComponent);
app.mount("#app");
