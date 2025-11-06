<template>
  <v-btn
    color="primary"
    tabindex="0"
    variant="elevated"
    aria-label="Dialog Connectors Add"
    data-test="connector-add-btn"
    @click="showDialog = true"
  >
    Add Docker Connector
  </v-btn>

  <ConnectorForm
    v-model="showDialog"
    :is-editing="false"
    :store-method="addConnector"
    data-test="connector-form-component"
    @update="$emit('update')"
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import ConnectorForm from "./ConnectorForm.vue";
import useConnectorStore from "@/store/modules/connectors";
import { IConnectorPayload } from "@/interfaces/IConnector";

defineEmits(["update"]);
const showDialog = ref(false);
const { createConnector } = useConnectorStore();

const addConnector = async (payload: Omit<IConnectorPayload, "uid">) => {
  await createConnector(payload);
};
</script>
