<template>
  <v-btn
    color="primary"
    tabindex="0"
    variant="elevated"
    aria-label="Dialog Connectors Add"
    data-test="connector-add-btn"
    @click="toggleDialog"
  >
    Add Docker Connector
  </v-btn>

  <ConnectorForm
    :is-editing="false"
    :show-dialog="dialog"
    :store-method="addConnector"
    @close="toggleDialog"
    data-test="connector-form-component"
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import ConnectorForm from "./ConnectorForm.vue";
import { useStore } from "../../store";

const dialog = ref(false);
const store = useStore();
const emit = defineEmits(["update"]);

const addConnector = async (payload) => {
  await store.dispatch("connectors/post", payload);
  emit("update");
};

const toggleDialog = () => {
  dialog.value = !dialog.value;
};
</script>
