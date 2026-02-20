import { App, Plugin, reactive } from "vue";

const SnackbarInjectionKey = Symbol("snackbar");

type SnackbarType = "success" | "error" | "info" | "warning";

interface SnackbarState {
  message: string;
  type: SnackbarType;
  show: boolean;
}

interface ISnackbarPlugin {
  showSuccess: (msg: string) => void;
  showError: (msg: string) => void;
  showInfo: (msg: string) => void;
  showWarning: (msg: string) => void;
  getMessage: () => string;
  getType: () => SnackbarType;
  getShow: () => boolean;
}

const state = reactive<SnackbarState>({
  message: "",
  type: "info",
  show: false,
});

let hideTimeout: ReturnType<typeof setTimeout> | undefined;

const showSnackbar = (type: SnackbarType, message: string) => {
  state.message = message;
  state.type = type;
  state.show = true;

  clearTimeout(hideTimeout);
  hideTimeout = setTimeout(() => {
    state.show = false;
  }, 4000);
};

const plugin: ISnackbarPlugin = {
  showSuccess: (msg: string) => showSnackbar("success", msg),
  showError: (msg: string) => showSnackbar("error", msg),
  showInfo: (msg: string) => showSnackbar("info", msg),
  showWarning: (msg: string) => showSnackbar("warning", msg),
  getMessage: () => state.message,
  getType: () => state.type,
  getShow: () => state.show,
};

const SnackbarPlugin: Plugin = {
  install(app: App) {
    app.provide(SnackbarInjectionKey, plugin);
  },
};

export { SnackbarPlugin, SnackbarInjectionKey, plugin };
export type { SnackbarState, ISnackbarPlugin };
