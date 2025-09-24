<template>
  <v-list-item @click="showDialog = true" data-test="privatekey-delete-btn">
    <div class="d-flex align-center">
      <div data-test="privatekey-delete-icon" class="mr-2">
        <v-icon> mdi-delete </v-icon>
      </div>

      <v-list-item-title data-test="privatekey-delete-btn-title">
        Remove
      </v-list-item-title>
    </div>
  </v-list-item>

  <MessageDialog
    v-model="showDialog"
    @close="showDialog = false"
    @confirm="remove"
    @cancel="showDialog = false"
    title="Are you sure?"
    description="You are about to delete this private key"
    icon="mdi-alert"
    icon-color="error"
    confirm-text="Delete"
    confirm-color="error"
    cancel-text="Close"
    confirm-data-test="confirm-btn"
    cancel-data-test="close-btn"
    data-test="private-key-delete-dialog"
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import useSnackbar from "@/helpers/snackbar";
import MessageDialog from "../MessageDialog.vue";
import usePrivateKeysStore from "@/store/modules/private_keys";

const props = defineProps<{ id: number }>();
const emit = defineEmits(["update"]);

const snackbar = useSnackbar();
const showDialog = ref(false);
const privateKeysStore = usePrivateKeysStore();

const remove = async () => {
  try {
    await privateKeysStore.deletePrivateKey(props.id);
    snackbar.showSuccess("The private key was removed successfully");
    emit("update");
  } finally {
    showDialog.value = false;
  }
};
</script>
