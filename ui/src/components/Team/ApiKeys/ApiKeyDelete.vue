<template>
  <v-list-item @click="showDialog = true" :disabled="!hasAuthorization">
    <div class="d-flex align-center">

      <div class="d-flex align-center">
        <div class="mr-2" data-test="delete-icon">
          <v-icon>mdi-delete</v-icon>
        </div>

        <v-list-item-title data-test="delete-main-btn-title"> Delete </v-list-item-title>
      </div>
    </div>
  </v-list-item>

  <BaseDialog v-model="showDialog">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-5 bg-primary" data-test="title">
        Are you sure?
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-4 mb-0 pb-1" data-test="text">
        <p class="text-body-2 mb-2">
          You are about to remove this Api Key from the namespace.
        </p>

        <p class="text-body-2 mb-2">
          After confirming this action cannot be redone.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" @click="showDialog = false" data-test="close-btn"> Close </v-btn>

        <v-btn color="red darken-1" variant="text" @click="remove()" data-test="delete-btn">
          Delete Key
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "@/components/BaseDialog.vue";
import useApiKeysStore from "@/store/modules/api_keys";

const props = defineProps<{
  keyId: string;
  hasAuthorization: boolean;
}>();

const snackbar = useSnackbar();
const emit = defineEmits(["update"]);
const showDialog = ref(false);
const apiKeyStore = useApiKeysStore();

const update = () => {
  emit("update");
  showDialog.value = false;
};

const remove = async () => {
  try {
    await apiKeyStore.removeApiKey({
      key: props.keyId,
    });
    update();
    snackbar.showSuccess("Api Key deleted successfully.");
  } catch (error: unknown) {
    snackbar.showError("Failed to delete Api Key.");
    handleError(error);
  }
};
</script>
