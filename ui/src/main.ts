import { createApp } from "vue";
import { createPinia } from "pinia";
import * as Sentry from "@sentry/vue";
import { BrowserTracing } from "@sentry/vue";
import { configure as configGtag } from "vue-gtag";
import { createChatWoot } from "@productdevbook/chatwoot/vue";
import { envVariables } from "./envVariables";
import vuetify from "./plugins/vuetify";
import { router } from "./router";
import App from "./App.vue";
import "asciinema-player/dist/bundle/asciinema-player.css";
import "@/nodespecific";
import loadFonts from "./plugins/webfontloader";
import "@fontsource/fira-code";
import "@fontsource/source-code-pro";
import "@fontsource/jetbrains-mono";
import "@fontsource/ubuntu-mono";
import "@fontsource/noto-mono";
import "@fontsource/inconsolata";
import "@fontsource/anonymous-pro";
import { SnackbarPlugin } from "./plugins/snackbar";

const pinia = createPinia();
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
app.use(pinia);

configGtag({
  tagId: envVariables.googleAnalyticsID || "",
});

if ((envVariables.isCloud) && (envVariables.chatWootWebsiteToken && envVariables.chatWootBaseURL)) {
  app.use(
    createChatWoot({
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
    }),
  );
}

app.use(SnackbarPlugin);
app.mount("#app");
