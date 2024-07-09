<template>
  <v-list-item v-bind="$attrs" @click="toggleDialog" :disabled="notHasAuthorization" data-test="connector-edit-btn">
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
    :is-editing="true"
    :initialAddress="props.ipAddress"
    :initialPort="props.portAddress"
    :initialSecure="props.secure"
    :uid="props.uid"
    :show-dialog="dialog"
    :store-method="editConnector"
    @close="toggleDialog"
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import ConnectorForm from "./ConnectorForm.vue";
import { useStore } from "../../store";

const props = defineProps({
  secure: {
    type: Boolean,
    required: true,
  },
  uid: {
    type: String,
    required: true,
  },
  ipAddress: {
    type: String,
    required: true,
  },
  portAddress: {
    type: Number,
    required: true,
  },
  notHasAuthorization: {
    type: Boolean,
    required: true,
    default: false,
  },
});

const emit = defineEmits(["update"]);
const dialog = ref(false);
const store = useStore();

const editConnector = async (payload) => {
  await store.dispatch("connectors/edit", payload);
  emit("update");
};

const toggleDialog = () => {
  dialog.value = !dialog.value;
};
</script>
