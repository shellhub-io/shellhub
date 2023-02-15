<template>
  <v-list-item
    v-bind="$attrs"
    @click="showDialog = true"
    :disabled="notHasAuthorization"
  >
    <div class="d-flex align-center">
      <div class="mr-2" data-test="remove-icon">
        <v-icon>mdi-delete</v-icon>
      </div>

      <v-list-item-title data-test="remove-title"> Remove </v-list-item-title>
    </div>
  </v-list-item>

  <v-dialog max-width="450" v-model="showDialog">
    <v-card class="bg-v-theme-surface" data-test="deviceDelete-card">
      <v-card-title class="text-h5 pa-5 bg-primary" data-test="text-title">
        Are you sure?
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-4 mb-0 pb-1" data-test="text-text">
        <p class="text-body-2 mb-2">You are about to remove this device.</p>

        <p class="text-body-2 mb-2">
          After confirming this action cannot be redone.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" @click="showDialog = false" data-test="close-btn">
          Close
        </v-btn>

        <v-btn
          color="red darken-1"
          variant="text"
          @click="remove()"
          data-test="remove-btn"
        >
          Remove
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { defineComponent, ref } from "vue";
import { useRouter } from "vue-router";
import {
  INotificationsError,
  INotificationsSuccess,
} from "@/interfaces/INotifications";
import { useStore } from "../../store";

export default defineComponent({
  props: {
    uid: {
      type: String,
      required: true,
    },

    redirect: {
      type: Boolean,
      default: false,
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
    const router = useRouter();

    const remove = async () => {
      try {
        await store.dispatch("devices/remove", props.uid);

        if (props.redirect) {
          router.push("/devices");
        } else {
          await store.dispatch("tags/fetch");
        }

        store.dispatch(
          "snackbar/showSnackbarSuccessAction",
          INotificationsSuccess.deviceDelete,
        );
        ctx.emit("update");
      } catch (error: any) {
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.deviceDelete,
        );
        throw new Error(error);
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
