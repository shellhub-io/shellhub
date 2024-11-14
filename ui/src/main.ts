import { createApp } from "vue";
import * as Sentry from "@sentry/vue";
import { BrowserTracing } from "@sentry/vue";
import VueGtag from "vue-gtag";
import { createChatWoot } from "@productdevbook/chatwoot/vue";
import { envVariables } from "./envVariables";
import vuetify from "./plugins/vuetify";
import { key, store } from "./store";
import { router } from "./router";
import App from "./App.vue";

import { loadFonts } from "./plugins/webfontloader";

import SnackbarComponent from "./components/Snackbar/Snackbar.vue";
import { SnackbarPlugin } from "./plugins/snackbar";

/* import the fontawesome core */

/* import font awesome icon component */

const app = createApp(App);

const chatwoot = createChatWoot({
  init: {
    websiteToken: envVariables.chatWootWebsiteToken,
    baseUrl: envVariables.chatWootBaseURL,
  },
  settings: {
    locale: "en",
    position: "right",
    hideMessageBubble: true,
  },
  partytown: false,
});

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
if (envVariables.isCloud || envVariables.isEnterprise) {
  app.use(chatwoot);
}
app.use(SnackbarPlugin);
app.component("SnackbarComponent", SnackbarComponent);
app.mount("#app");
