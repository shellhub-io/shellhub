<template>
  <v-list-item v-bind="$attrs" @click="showDialog = true" :disabled="!hasAuthorization" data-test="connector-remove-btn">
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon data-test="remove-icon"> mdi-delete </v-icon>
      </div>

      <v-list-item-title data-test="remove-title"> Remove </v-list-item-title>
    </div>
  </v-list-item>

  <BaseDialog v-model="showDialog">
    <v-card class="bg-v-theme-surface">
      <v-card-title data-test="text-title" class="text-h5 pa-5 bg-primary">
        Are you sure?
      </v-card-title>
      <v-divider />

      <v-card-text data-test="text" class="mt-4 mb-0 pb-1">
        <p class="text-body-2 mb-2">You are about to remove this Docker Connector</p>

        <p class="text-body-2 mb-2">
          After confirming this action cannot be redone.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn data-test="close-btn" variant="text" @click="showDialog = false">
          Close
        </v-btn>

        <v-btn
          data-test="remove-btn"
          color="red darken-1"
          variant="text"
          @click="remove()"
        >
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

const props = defineProps<{
  uid: string;
  hasAuthorization: boolean;
}>();

const emit = defineEmits(["update"]);
const showDialog = ref(false);
const store = useStore();
const snackbar = useSnackbar();
const remove = async () => {
  try {
    await store.dispatch("connectors/remove", props.uid);
    snackbar.showSuccess("Successfully removed connector.");
    showDialog.value = false;
    emit("update");
  } catch (error: unknown) {
    snackbar.showError("Failed to remove connector.");
    handleError(error);
  }
};

</script>
