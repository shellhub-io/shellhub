// src/global-components.d.ts
import SnackbarComponent from "./components/Snackbar/Snackbar.vue";

declare module "@vue/runtime-core" {
  export interface GlobalComponents {
    SnackbarComponent: typeof SnackbarComponent;
    RouterLink: typeof import("vue-router")["RouterLink"];
    RouterView: typeof import("vue-router")["RouterView"];
  }
}
