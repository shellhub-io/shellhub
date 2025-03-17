<template>
  <v-list-item v-bind="$attrs" @click="showDialog = true" :disabled="notHasAuthorization" data-test="open-tag-remove">
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon> mdi-delete </v-icon>
      </div>

      <v-list-item-title data-test="mdi-information-list-item">
        Remove
      </v-list-item-title>
    </div>
  </v-list-item>

  <v-dialog max-width="450" v-model="showDialog">
    <v-card class="bg-v-theme-surface" data-test="dialog-card">
      <v-card-title class="text-h5 pa-5 bg-primary" data-test="dialog-title">
        Are you sure?
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-4 mb-0 pb-1" data-test="dialog-subtitle">
        <p class="text-body-2 mb-2">You are about to remove this tag.</p>

        <p class="text-body-2 mb-2">
          After confirming this action cannot be redone.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" @click="close()" data-test="close-btn"> Close </v-btn>

        <v-btn color="red darken-1" variant="text" @click="remove()" data-test="remove-btn">
          Remove
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import { useStore } from "../../store";
import handleError from "@/utils/handleError";

const props = defineProps({
  tagName: {
    type: String,
    required: true,
  },
  notHasAuthorization: {
    type: Boolean,
    required: true,
  },
});

const emit = defineEmits(["update"]);
const tenant = computed(() => localStorage.getItem("tenant"));

const showDialog = ref(false);
const store = useStore();

const close = () => {
  showDialog.value = false;
};

const update = () => {
  emit("update");
  close();
};

const remove = async () => {
  try {
    await store.dispatch("tags/removeTag", {
      currentName: props.tagName,
      tenant: tenant.value,
    });

    update();

    store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      INotificationsSuccess.deviceTagDelete,
    );
  } catch (error: unknown) {
    store.dispatch(
      "snackbar/showSnackbarErrorAction",
      INotificationsError.deviceTagDelete,
    );
    handleError(error);
  }
};
</script>
