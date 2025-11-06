<template>
  <v-list-item
    data-test="privatekey-delete-btn"
    @click="showDialog = true"
  >
    <div class="d-flex align-center">
      <div
        data-test="privatekey-delete-icon"
        class="mr-2"
      >
        <v-icon> mdi-delete </v-icon>
      </div>

      <v-list-item-title data-test="privatekey-delete-btn-title">
        Remove
      </v-list-item-title>
    </div>
  </v-list-item>

  <MessageDialog
    v-model="showDialog"
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
    @close="showDialog = false"
    @confirm="remove"
    @cancel="showDialog = false"
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import useSnackbar from "@/helpers/snackbar";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
import usePrivateKeysStore from "@/store/modules/private_keys";

const props = defineProps<{ id: number }>();
const emit = defineEmits(["update"]);

const snackbar = useSnackbar();
const showDialog = ref(false);
const privateKeysStore = usePrivateKeysStore();

const remove = () => {
  try {
    privateKeysStore.deletePrivateKey(props.id);
    snackbar.showSuccess("The private key was removed successfully");
    emit("update");
  } finally {
    showDialog.value = false;
  }
};
</script>
