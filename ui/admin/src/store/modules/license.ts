import { defineStore } from "pinia";
import { GetLicense200Response } from "@admin/api/client/api";
import * as apiLicense from "../api/license";

export const useLicenseStore = defineStore("license", {
  state: (): { license: GetLicense200Response } => ({
    license: {} as GetLicense200Response,
  }),

  getters: {
    isExpired: (state): boolean => (state.license && state.license.expired)
      || (state.license && state.license.expired === undefined),

    getLicense: (state): GetLicense200Response => state.license,
  },

  actions: {
    async get() {
      const res = await apiLicense.getLicense();
      this.license = res.data;
    },

    async post(file: File) {
      await apiLicense.uploadLicense(file);
    },
  },
});

export default useLicenseStore;
