import { configuration } from "@/api/http";

const apiPlugin = (store) => {
  store.subscribe((mutation, state) => {
    if (mutation.type === "auth/authSuccess") {
      configuration.accessToken = state.auth.token;
    }
  });
};

export default apiPlugin;
