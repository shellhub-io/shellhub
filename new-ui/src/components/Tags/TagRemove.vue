<template>
  <v-list-item @click="showDialog = true" :disabled="notHasAuthorization">
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon color="white"> mdi-delete </v-icon>
      </div>

      <v-list-item-title data-test="mdi-information-list-item">
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
  </v-dialog>
</template>

<script lang="ts">
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import { defineComponent, ref } from "vue";
import { useStore } from "../../store";

export default defineComponent({
  props: {
    tagName: {
      type: String,
      required: true,
    },
    notHasAuthorization: {
      type: Boolean,
      required: true,
    },
  },
  emits: ["update"],
  setup(props, ctx) {
    const showDialog = ref(false);
    const store = useStore();

    const remove = async () => {
      try {
        await store.dispatch("tags/remove", props.tagName);

        store.dispatch(
          "snackbar/showSnackbarSuccessAction",
          INotificationsSuccess.deviceTagDelete
        );
        ctx.emit("update");
      } catch {
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.deviceTagDelete
        );
      } finally {
        showDialog.value = false;
      }
    };

    return {
      showDialog,
      remove,
    };
  },
});
</script>
