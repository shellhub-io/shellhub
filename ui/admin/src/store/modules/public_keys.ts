import { defineStore } from "pinia";
import {
  postPublicKey,
  fetchPublicKeys,
  getPublicKey,
  putPublicKey,
  removePublicKey,
} from "../api/public_keys";
import { IAdminPublicKey } from "../../interfaces/IPublicKey";

export const usePublicKeysStore = defineStore("publicKeys", {
  state: () => ({
    publicKeys: [] as Array<IAdminPublicKey>,
    publicKey: {} as IAdminPublicKey,
    numberPublicKeys: 0,
  }),

  getters: {
    list: (state) => state.publicKeys,
    get: (state) => state.publicKey,
    getNumberPublicKeys: (state) => state.numberPublicKeys,
  },

  actions: {
    async post(data) {
      await postPublicKey(data);
    },

    async fetch(data: { perPage: number; page: number }) {
      try {
        const res = await fetchPublicKeys(data.perPage, data.page);
        this.publicKeys = res.data;
        this.numberPublicKeys = parseInt(res.headers["x-total-count"], 10);
      } catch (error) {
        this.clearListPublicKeys();
        throw error;
      }
    },

    async get(id: string) {
      try {
        const res = await getPublicKey(id);
        this.publicKey = res.data;
      } catch (error) {
        this.clearObjectPublicKey();
        throw error;
      }
    },

    async put(data) {
      await putPublicKey(data);
    },

    async remove(fingerprint: string) {
      await removePublicKey(fingerprint);

      const index = this.publicKeys.findIndex(
        (d: IAdminPublicKey) => d.fingerprint === fingerprint,
      );

      if (index !== -1) {
        this.publicKeys.splice(index, 1);
      }
    },

    clearListPublicKeys() {
      this.publicKeys = [];
      this.numberPublicKeys = 0;
    },

    clearObjectPublicKey() {
      this.publicKey = {} as IAdminPublicKey;
    },
  },
});

export default usePublicKeysStore;
