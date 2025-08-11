<template>
  <v-list-item v-bind="$attrs" @click="showDialog = true" :disabled="!hasAuthorization" data-test="connector-edit-btn">
    <div class="d-flex align-center">
      <div data-test="connector-edit-icon" class="mr-2">
        <v-icon> mdi-pencil </v-icon>
      </div>

      <v-list-item-title>
        Edit
      </v-list-item-title>
    </div>
  </v-list-item>

  <ConnectorForm
    v-model="showDialog"
    :is-editing="true"
    :initialAddress="props.ipAddress"
    :initialPort="props.portAddress"
    :initialSecure="props.secure"
    :uid="props.uid"
    :store-method="editConnector"
    @update="$emit('update')"
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import ConnectorForm from "./ConnectorForm.vue";
import useConnectorStore from "@/store/modules/connectors";

const props = defineProps<{
  secure: boolean;
  uid: string;
  ipAddress: string;
  portAddress: number;
  hasAuthorization: boolean;
}>();

defineEmits(["update"]);
const showDialog = ref(false);
const { updateConnector } = useConnectorStore();

const editConnector = async (payload) => {
  await updateConnector(payload);
};
</script>
