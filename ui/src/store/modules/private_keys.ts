import { defineStore } from "pinia";
import { ref } from "vue";
import { IPrivateKey } from "@/interfaces/IPrivateKey";

const usePrivateKeysStore = defineStore("privateKey", () => {
  const privateKeys = ref<Array<IPrivateKey>>([]);

  const getPrivateKeyList = () => {
    privateKeys.value = JSON.parse(localStorage.getItem("privateKeys") || "[]");
  };

  const validateKeyUniqueness = (name: string, data: string, currentKeyId?: number) => {
    // currentKeyId prevents validating against the key being edited
    const hasSameName = privateKeys.value.some((key) => key.id !== currentKeyId && key.name === name);
    const hasSameData = privateKeys.value.some((key) => key.id !== currentKeyId && key.data === data);

    if (hasSameName && hasSameData) throw new Error("both");
    if (hasSameData) throw new Error("private_key");
    if (hasSameName) throw new Error("name");
  };

  const addPrivateKey = (newKey: Omit<IPrivateKey, "id">) => {
    validateKeyUniqueness(newKey.name, newKey.data);

    const existentIds = privateKeys.value.map((key) => key.id);
    const newKeyId = existentIds.length ? Math.max(...existentIds) + 1 : 1;

    privateKeys.value.push({ ...newKey, id: newKeyId });
    localStorage.setItem("privateKeys", JSON.stringify(privateKeys.value));
  };

  const editPrivateKey = (updatedKey: IPrivateKey) => {
    const index = privateKeys.value.findIndex((key) => key.id === updatedKey.id);
    if (index === -1) throw new Error("Key not found");

    const existingKey = privateKeys.value[index];

    if (existingKey.name !== updatedKey.name || existingKey.data !== updatedKey.data) {
      validateKeyUniqueness(updatedKey.name, updatedKey.data, updatedKey.id);
    }

    privateKeys.value.splice(index, 1, updatedKey);
    localStorage.setItem("privateKeys", JSON.stringify(privateKeys.value));
  };

  const deletePrivateKey = (id: number) => {
    privateKeys.value = privateKeys.value.filter((key) => key.id !== id);
    localStorage.setItem("privateKeys", JSON.stringify(privateKeys.value));
  };

  return {
    privateKeys,

    getPrivateKeyList,
    addPrivateKey,
    editPrivateKey,
    deletePrivateKey,
  };
});

export default usePrivateKeysStore;
