<template>
  <v-list-item v-bind="$attrs" @click="showDialog = true" :disabled="!hasAuthorization">
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon> mdi-delete </v-icon>
      </div>

      <v-list-item-title data-test="mdi-information-list-item">
        Remove
      </v-list-item-title>
    </div>
  </v-list-item>

  <BaseDialog v-model="showDialog">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-5 bg-primary">
        Are you sure?
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-4 mb-0 pb-1">
        <p class="text-body-2 mb-2">You are about to remove this tag.</p>

        <p class="text-body-2 mb-2">
          After confirming this action cannot be redone.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" @click="showDialog = false"> Close </v-btn>

        <v-btn color="red darken-1" variant="text" @click="remove()">
          Remove
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";

defineOptions({
  inheritAttrs: false,
});

const props = defineProps<{
  tag: string;
  hasAuthorization: boolean;
}>();

const emit = defineEmits(["update"]);
const showDialog = ref(false);
const store = useStore();
const snackbar = useSnackbar();

const remove = async () => {
  try {
    await store.dispatch("tags/remove", props.tag);
    snackbar.showSuccess(`Tag ${props.tag} removed successfully.`);
    emit("update");
  } catch (error: unknown) {
    snackbar.showError("Failed to remove tag.");
    handleError(error);
  } finally {
    showDialog.value = false;
  }
};
</script>
