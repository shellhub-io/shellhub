import { configuration, reloadConfiguration } from "@/api/http";

const apiPlugin = (store) => {
  store.subscribe((mutation, state) => {
    if (mutation.type === "auth/authSuccess") {
      configuration.accessToken = state.auth.token;
      reloadConfiguration();
    }
  });
};

export default apiPlugin;
