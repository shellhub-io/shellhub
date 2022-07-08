<template>
  <v-list-item
    @click="showDialog = true"
    v-bind="$props"
    :disabled="notHasAuthorization"
  >
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon color="white"> mdi-close-circle </v-icon>
      </div>

      <v-list-item-title data-test="mdi-information-list-item">
        Close Session
      </v-list-item-title>
    </div>
  </v-list-item>

  <v-dialog max-width="500" v-model="showDialog">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-3 bg-primary">
        Are you sure?
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-4 mb-0 pb-1">
        <p class="text-body-2 mb-2">
          You are going to close connection for this device.
        </p>

        <p class="text-body-2 mb-2">
          After confirming this action cannot be redone.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" @click="showDialog = false"> Cancel </v-btn>

        <v-btn color="red darken-1" variant="text" @click="closeSession()">
          Close
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
import { defineComponent, PropType, ref } from "vue";
import { IDevice } from "../../interfaces/IDevice";
import { useStore } from "../../store";

export default defineComponent({
  props: {
    uid: {
      type: String,
      required: true,
    },
    device: {
      type: Object as PropType<IDevice>,
      required: true,
    },
    notHasAuthorization: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["update"],
  setup(props, ctx) {
    const showDialog = ref(false);
    const store = useStore();

    const closeSession = async () => {
      try {
        await store.dispatch("sessions/close", {
          uid: props.uid,
          device_uid: props.device.uid,
        });
        showDialog.value = false;
        store.dispatch(
          "snackbar/showSnackbarSuccessAction",
          INotificationsSuccess.sessionClose
        );
        ctx.emit("update");
      } catch {
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.sessionClose
        );
      }
    };

    return {
      showDialog,
      closeSession,
    };
  },
});
</script>
