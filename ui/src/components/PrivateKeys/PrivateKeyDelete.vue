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

  <BaseDialog v-model="showDialog">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-5 bg-primary" data-test="privatekey-dialog-title">
        Are you sure?
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-4 mb-0 pb-1" data-test="privatekey-dialog-text">
        <p class="text-body-2 mb-2">
          You are about to remove this private key.
        </p>

        <p class="text-body-2 mb-2">
          After confirming this action cannot be redone.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" @click="showDialog = false" data-test="privatekey-close-btn"> Close </v-btn>

        <v-btn color="red darken-1" variant="text" @click="remove()" data-test="privatekey-remove-btn">
          Remove
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
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
