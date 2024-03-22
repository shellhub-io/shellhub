<template>
  <v-list-item v-bind="$attrs" @click="showDialog = true" :disabled="notHasAuthorization" data-test="public-key-remove-btn">
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon data-test="remove-icon"> mdi-delete </v-icon>
      </div>

      <v-list-item-title data-test="remove-title"> Remove </v-list-item-title>
    </div>
  </v-list-item>

  <v-dialog max-width="450" v-model="showDialog">
    <v-card class="bg-v-theme-surface">
      <v-card-title data-test="text-title" class="text-h5 pa-5 bg-primary">
        Are you sure?
      </v-card-title>
      <v-divider />

      <v-card-text data-test="text" class="mt-4 mb-0 pb-1">
        <p class="text-body-2 mb-2">You are about to remove this public key.</p>

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
  fingerprint: {
    type: String,
    required: true,
  },
  notHasAuthorization: {
    type: Boolean,
    required: true,
  },
});
const emit = defineEmits(["update"]);
const showDialog = ref(false);
const store = useStore();
const remove = async () => {
  try {
    await store.dispatch("publicKeys/remove", props.fingerprint);

    store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      INotificationsSuccess.publicKeyDeleting,
    );
    emit("update");
  } catch (error: unknown) {
    store.dispatch(
      "snackbar/showSnackbarErrorAction",
      INotificationsError.publicKeyDeleting,
    );
    handleError(error);
  } finally {
    showDialog.value = false;
  }
};

</script>
