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
    @update="$emit('update')"
    data-test="connector-form-component"
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import ConnectorForm from "./ConnectorForm.vue";
import { useStore } from "@/store";

defineEmits(["update"]);
const showDialog = ref(false);
const store = useStore();

const addConnector = async (payload) => {
  await store.dispatch("connectors/post", payload);
};
</script>
