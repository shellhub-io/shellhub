import { Module } from "vuex";
import * as apiKeysApi from "../api/api_keys";
import { ApiKey } from "@/interfaces/IUserLogin";
import { State } from "..";

export interface ApiKeysState {
  page: number;
  perPage: number;
  sortStatusField: undefined | string;
  sortStatusString: "asc" | "desc" | "";
  keyList: Array<ApiKey>,
  keyResponse: string,
  numberApiKeys: number,
}
export const apiKeys: Module<ApiKeysState, State> = {
  namespaced: true,
  state: {
    page: 1,
    perPage: 10,
    sortStatusField: undefined,
    sortStatusString: "asc",
    keyList: [],
    keyResponse: "",
    numberApiKeys: 0,
  },

  getters: {
    getSortStatusField: (state) => state.sortStatusField,
    getSortStatusString: (state) => state.sortStatusString,
    apiKey: (state) => state.keyResponse,
    apiKeyList: (state) => state.keyList,
    getNumberApiKeys: (state) => state.numberApiKeys,
  },

  mutations: {
    clearApiKeysList: (state) => {
      state.keyList = [];
      state.numberApiKeys = 0;
    },

    setQueryApiKeyGet: (state, data) => {
      state.page = data.page;
      state.perPage = data.perPage;
      state.sortStatusField = data.sortStatusField;
      state.sortStatusString = data.sortStatusString;
    },

    apiKey(state, data) {
      state.keyResponse = data.id;
    },

    setKeyList(state, res) {
      state.keyList = res.data;
      state.numberApiKeys = parseInt(res.headers["x-total-count"], 10);
    },

    setSortStatus: (state, data) => {
      state.sortStatusString = data.sortStatusString;
      state.sortStatusField = data.sortStatusField;
    },
  },

  actions: {
    async generateApiKey(context, data) {
      try {
        const resp = await apiKeysApi.generateApiKey(data);
        if (resp.status === 200) {
          context.commit("apiKey", resp.data);
        }
      } catch (error) {
        context.commit("authError");
        throw error;
      }
    },

    async getApiKey(context, data) {
      try {
        const resp = await apiKeysApi.getApiKey(
          data.page,
          data.perPage,
          data.sortStatusString,
          data.sortStatusField,
        );
        if (resp.data.length) {
          context.commit("setKeyList", resp);
          context.commit("setQueryApiKeyGet", data);
          return resp;
        }

        context.commit("clearApiKeysList");
        return false;
      } catch (error) {
        context.commit("authError");
        throw error;
      }
    },

    async editApiKey(context, data) {
      try {
        await apiKeysApi.editApiKey(data);
      } catch (error) {
        context.commit("authError");
        throw error;
      }
    },

    async removeApiKey(context, data) {
      try {
        await apiKeysApi.removeApiKey(data);
      } catch (error) {
        context.commit("authError");
        throw error;
      }
    },

    async setSortStatus({ commit }, data) {
      commit("setSortStatus", data);
    },
  },
};
