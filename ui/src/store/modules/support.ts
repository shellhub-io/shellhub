import { defineStore } from "pinia";
import { ref } from "vue";
import { useChatWoot } from "@productdevbook/chatwoot/vue";
import * as supportApi from "../api/namespaces";

const useSupportStore = defineStore("support", () => {
  const identifier = ref<string>("");
  const isChatCreated = ref<boolean>(false);

  const getIdentifier = async (tenantId: string) => {
    useChatWoot().reset();
    const res = await supportApi.getSupportID(tenantId);
    identifier.value = res.data.identifier as string;
  };

  return {
    identifier,
    isChatCreated,
    getIdentifier,
  };
});

export default useSupportStore;
