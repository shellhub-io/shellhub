<template>
  <v-list-item v-bind="$attrs" @click="showDialog = true" :disabled="notHasAuthorization" data-test="connector-edit-btn">
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
    v-model="showDialog"
    :store-method="editConnector"
    @update="$emit('update')"
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import ConnectorForm from "./ConnectorForm.vue";
import { useStore } from "@/store";

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
    default: false,
  },
});

defineEmits(["update"]);
const showDialog = ref(false);
const store = useStore();

const editConnector = async (payload) => {
  await store.dispatch("connectors/edit", payload);
};
</script>
