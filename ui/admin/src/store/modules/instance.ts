// stores/instance.ts
import { defineStore } from "pinia";
import * as apiInstance from "../api/instance";
import { IAdminSAMLConfig } from "../../interfaces/IInstance";

export interface InstanceState {
  authenticationSettings: {
    local?: {
      enabled: boolean;
    };
    saml?: {
      enabled: boolean;
      auth_url: string;
      assertion_url: string;
      idp: {
        entity_id: string;
        signon_url: string;
        certificates: string[];
      };
      sp: {
        sign_requests: boolean;
        certificate?: string;
      };
    };
  };
}

export const useInstanceStore = defineStore("instance", {
  state: (): InstanceState => ({
    authenticationSettings: {
      local: {
        enabled: false,
      },
      saml: {
        enabled: false,
        auth_url: "",
        assertion_url: "",
        idp: {
          entity_id: "",
          signon_url: "",
          certificates: [],
        },
        sp: {
          sign_requests: false,
          certificate: "",
        },
      },
    },
  }),

  getters: {
    getAuthenticationSettings: (state) => state.authenticationSettings,
    isLocalAuthEnabled: (state) => state.authenticationSettings?.local?.enabled,
    isSamlEnabled: (state) => state.authenticationSettings?.saml?.enabled,

  },

  actions: {
    async fetchAuthenticationSettings() {
      const response = await apiInstance.getAuthenticationSettings();
      this.authenticationSettings = response.data as never;
    },

    async updateLocalAuthentication(status: boolean) {
      await apiInstance.configureLocalAuthentication(status);
      await this.fetchAuthenticationSettings();
    },

    async updateSamlAuthentication(data: IAdminSAMLConfig) {
      await apiInstance.configureSAMLAuthentication(data);
      await this.fetchAuthenticationSettings();
    },
  },
});

export default useInstanceStore;
