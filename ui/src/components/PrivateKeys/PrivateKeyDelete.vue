<template>
  <v-list-item @click="showDialog = true">
    <div class="d-flex align-center">
      <div data-test="privatekey-delete-icon" class="mr-2">
        <v-icon> mdi-delete </v-icon>
      </div>

      <v-list-item-title data-test="privatekey-delete-title">
        Remove
      </v-list-item-title>
    </div>
  </v-list-item>

  <v-dialog max-width="450" v-model="showDialog">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-5 bg-primary">
        Are you sure?
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-4 mb-0 pb-1">
        <p class="text-body-2 mb-2">
          You are about to remove this private key.
        </p>

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
  </v-dialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import { useStore } from "../../store";
import handleError from "@/utils/handleError";

const props = defineProps({
  id: {
    type: Number,
    required: true,
  },
});
const emit = defineEmits(["update"]);
const showDialog = ref(false);
const store = useStore();

const remove = async () => {
  try {
    await store.dispatch("privateKey/remove", props.id);
    store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      INotificationsSuccess.privateKeyDeleting,
    );
    emit("update");
  } catch (error: unknown) {
    store.dispatch(
      "snackbar/showSnackbarErrorAction",
      INotificationsError.privateKeyDeleting,
    );
    handleError(error);
  } finally {
    showDialog.value = false;
  }
};
</script>
