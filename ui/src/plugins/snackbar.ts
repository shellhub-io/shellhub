import { App, Plugin } from "vue";

const InjectionKey = "snackbar";

interface PluginInterface {
  showInfo(message: string);
  showSuccess(message: string);
  showWarning(message: string);
  showError(message: string);
}

const SnackbarPlugin: Plugin = {
  install(app: App) {
    const store = app.config.globalProperties.$store;

    const plugin = {
      showInfo(message: string) {
        store.commit("snackbar/showMessage", { type: "info", message });
      },
      showSuccess(message: string) {
        store.commit("snackbar/showMessage", { type: "success", message });
      },
      showWarning(message: string) {
        store.commit("snackbar/showMessage", { type: "warning", message });
      },
      showError(message: string) {
        store.commit("snackbar/showMessage", { type: "error", message });
      },
    } as PluginInterface;

    app.provide(InjectionKey, plugin);
  },
};

export { InjectionKey, type PluginInterface, SnackbarPlugin };
