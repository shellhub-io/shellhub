import { createApp } from "vue";
import * as Sentry from "@sentry/vue";
import { BrowserTracing } from "@sentry/tracing";
import { envVariables } from "./envVariables";
import vuetify from "./plugins/vuetify";
import { key, store } from "./store";
import { router } from "./router";
import App from "./App.vue";

import { loadFonts } from "./plugins/webfontloader";

import SnackbarComponent from "./components/Snackbar/Snackbar.vue";
import { SnackbarPlugin } from "./plugins/snackbar";

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
app.use(SnackbarPlugin);
app.component("SnackbarComponent", SnackbarComponent);
app.mount("#app");
