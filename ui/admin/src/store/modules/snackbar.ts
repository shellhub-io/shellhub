// stores/snackbar.ts
import { defineStore } from "pinia";

export interface SnackbarMessageAndContentType {
  typeMessage: string
  typeContent: string
}

export interface SnackbarState {
  snackbarError: boolean
  snackbarSuccess: boolean
  snackbarMessageAndContentType: SnackbarMessageAndContentType
  snackbarCopy: boolean
}

export const useSnackbarStore = defineStore("snackbar", {
  state: (): SnackbarState => ({
    snackbarError: false,
    snackbarSuccess: false,
    snackbarMessageAndContentType: { typeMessage: "", typeContent: "" },
    snackbarCopy: false,
  }),

  getters: {
    getSnackbarSuccess: (state) => state.snackbarSuccess,
    getSnackbarError: (state) => state.snackbarError,
    getSnackbarMessageAndContentType: (state) => state.snackbarMessageAndContentType,
    getSnackbarCopy: (state) => state.snackbarCopy,
  },

  actions: {
    showSnackbarSuccessAction(value: string) {
      this.snackbarMessageAndContentType = { typeMessage: "action", typeContent: value };
      this.snackbarSuccess = true;
    },

    showSnackbarSuccessDefault() {
      this.snackbarMessageAndContentType = { typeMessage: "default", typeContent: "" };
      this.snackbarSuccess = true;
    },

    unsetShowStatusSnackbarSuccess() {
      this.snackbarSuccess = false;
    },

    showSnackbarErrorLoading(value: string) {
      this.snackbarMessageAndContentType = { typeMessage: "loading", typeContent: value };
      this.snackbarError = true;
    },

    showSnackbarNoContent() {
      this.snackbarMessageAndContentType = { typeMessage: "no-content", typeContent: "" };
      this.snackbarSuccess = true;
    },

    showSnackbarErrorAction(value: string) {
      this.snackbarMessageAndContentType = { typeMessage: "action", typeContent: value };
      this.snackbarError = true;
    },

    showSnackbarErrorCustom(value: string) {
      this.snackbarMessageAndContentType = { typeMessage: "custom", typeContent: value };
      this.snackbarError = true;
    },

    showSnackbarErrorLicense(value: string) {
      this.snackbarMessageAndContentType = { typeMessage: "licenseRequired", typeContent: value };
      this.snackbarError = true;
    },

    showSnackbarErrorDefault() {
      this.snackbarMessageAndContentType = { typeMessage: "default", typeContent: "" };
      this.snackbarError = true;
    },

    unsetShowStatusSnackbarError() {
      this.snackbarError = false;
    },

    showSnackbarCopy(value: string) {
      this.snackbarMessageAndContentType = { typeMessage: "", typeContent: value };
      this.snackbarCopy = true;
    },

    unsetShowStatusSnackbarCopy() {
      this.snackbarCopy = false;
    },
  },
});

export default useSnackbarStore;
