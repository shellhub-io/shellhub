import { Module } from "vuex";
import { State } from "./../index";
import * as apiInstance from "../api/instance";
import { IConfigureSAML } from "../../interfaces/IInstance";

export interface InstanceState {
  authenticationSettings: {
    local: {
      enabled: boolean;
    };
    saml: {
      enabled: boolean;
      auth_url: string;
      idp: {
        entity_id: string;
        signon_url: string;
        certificates: string[];
      };
      sp: {
        sign_requests: boolean;
      };
    };
  } | null;
}

export const instance: Module<InstanceState, State> = {
  namespaced: true,
  state: {
    authenticationSettings: {
      local: {
        enabled: false,
      },
      saml: {
        enabled: false,
        auth_url: "",
        idp: {
          entity_id: "",
          signon_url: "",
          certificates: [],
        },
        sp: {
          sign_requests: false,
        },
      },
    },
  },

  getters: {
    authenticationSettings: (state) => state.authenticationSettings,
    isLocalAuthEnabled: (state) => state.authenticationSettings?.local.enabled,
    isSamlEnabled: (state) => state.authenticationSettings?.saml.enabled,
  },

  mutations: {
    setAuthenticationSettings(state, settings) {
      state.authenticationSettings = settings;
    },
  },

  actions: {
    async fetchAuthenticationSettings({ commit }) {
      const response = await apiInstance.getAuthenticationSettings();
      await commit("setAuthenticationSettings", response.data);
    },

    async updateLocalAuthentication({ dispatch }, status: boolean) {
      await apiInstance.configureLocalAuthentication(status);
      await dispatch("fetchAuthenticationSettings");
    },

    async updateSamlAuthentication({ dispatch }, data: IConfigureSAML) {
      await apiInstance.configureSAMLAuthentication(data);
      await dispatch("fetchAuthenticationSettings");
    },
  },
};
