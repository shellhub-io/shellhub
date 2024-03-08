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

  <v-dialog max-width="450" v-model="showDialog">
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
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import { useStore } from "../../store";
import handleError from "@/utils/handleError";

const props = defineProps({
  keyId: {
    type: String,
    required: true,
  },
  hasAuthorization: {
    type: Boolean,
    required: true,
  },
});
const emit = defineEmits(["update"]);
const showDialog = ref(false);
const store = useStore();
const tenant = computed(() => localStorage.getItem("tenant"));

const update = () => {
  emit("update");
  showDialog.value = false;
};

const remove = async () => {
  try {
    await store.dispatch("auth/removeApiKey", {
      tenant: tenant.value,
      id: props.keyId,
    });
    update();
    store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      INotificationsSuccess.deleteKey,
    );
  } catch (error: unknown) {
    store.dispatch(
      "snackbar/showSnackbarErrorAction",
      INotificationsError.deleteKey,
    );
    handleError(error);
  }
};
</script>
