import { defineStore } from "pinia";
import { ref } from "vue";
import { IPrivateKey } from "@/interfaces/IPrivateKey";

const usePrivateKeysStore = defineStore("privateKey", () => {
  const privateKeys = ref<Array<IPrivateKey>>([]);

  const getPrivateKeyList = () => {
    privateKeys.value = JSON.parse(localStorage.getItem("privateKeys") || "[]");
  };

  const addPrivateKey = (newKey: Omit<IPrivateKey, "id">) => {
    const existentIds: number[] = [];

    privateKeys.value.forEach((key: IPrivateKey) => {
      if (key.data === newKey.data && key.name === newKey.name) throw new Error("both");
      if (key.data === newKey.data) throw new Error("private_key");
      if (key.name === newKey.name) throw new Error("name");
      existentIds.push(key.id);
    });

    const newKeyId = existentIds.length ? Math.max(...existentIds) + 1 : 1;
    privateKeys.value.push({ ...newKey, id: newKeyId });
    localStorage.setItem("privateKeys", JSON.stringify(privateKeys.value));
  };

  const editPrivateKey = (updatedKey: IPrivateKey) => {
    const index = privateKeys.value.findIndex((key) => key.id === updatedKey.id);
    const existingKey = privateKeys.value[index];

    if (existingKey && existingKey.data === updatedKey.data && existingKey.name === updatedKey.name) {
      throw new Error();
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
