<template>
  <v-btn
    color="primary"
    tabindex="0"
    variant="elevated"
    aria-label="Dialog Connectors Add"
    data-test="connector-add-btn"
    @click="openDialog"
  >
    Add Connector
  </v-btn>

  <ConnectorForm
    :is-editing="false"
    :show-dialog="dialog"
    :store-method="addConnector"
    @close="closeDialog"
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import ConnectorForm from "./ConnectorForm.vue";
import { useStore } from "../../store";

const dialog = ref(false);
const store = useStore();

const addConnector = async (payload) => {
  await store.dispatch("connectors/post", payload);
};

const openDialog = () => {
  dialog.value = true;
};

const closeDialog = () => {
  dialog.value = false;
};
</script>
