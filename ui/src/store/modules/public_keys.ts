import { defineStore } from "pinia";
import { ref } from "vue";
import * as publicKeysApi from "../api/public_keys";
import { IPublicKey, IPublicKeyCreate } from "@/interfaces/IPublicKey";
import { parseTotalCount } from "@/utils/headers";

const usePublicKeysStore = defineStore("publicKeys", () => {
  const publicKeys = ref<Array<IPublicKey>>([]);
  const publicKeyCount = ref(0);

  const fetchPublicKeyList = async (data?: { page: number; perPage: number; filter?: string }) => {
    const res = await publicKeysApi.fetchPublicKeys(data?.page || 1, data?.perPage || 10, data?.filter);
    publicKeys.value = res.data as IPublicKey[];
    publicKeyCount.value = parseTotalCount(res.headers);
  };

  const createPublicKey = async (data: IPublicKeyCreate) => {
    await publicKeysApi.createPublicKey(data);
  };

  const updatePublicKey = async (data: IPublicKey) => {
    await publicKeysApi.updatePublicKey(data);
  };

  const removeFromList = (fingerprint: string) => {
    const index = publicKeys.value.findIndex((d) => d.fingerprint === fingerprint);
    if (index !== -1) {
      publicKeys.value.splice(index, 1);
      publicKeyCount.value--;
    }
  };

  const deletePublicKey = async (fingerprint: string) => {
    await publicKeysApi.deletePublicKey(fingerprint);
    removeFromList(fingerprint);
  };

  return {
    publicKeys,
    publicKeyCount,

    createPublicKey,
    fetchPublicKeyList,
    updatePublicKey,
    deletePublicKey,
  };
});

export default usePublicKeysStore;
