<template>
  <v-list-item @click="showDialog = true" :disabled="!hasAuthorization">
    <div class="d-flex align-center">
      <div data-test="namespace-delete-icon" class="mr-2">
        <v-icon> mdi-delete </v-icon>
      </div>

      <v-list-item-title data-test="namespace-delete-title">
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
          You are about to remove this user from the namespace.
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

<script lang="ts">
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import { defineComponent, ref } from "vue";
import { useStore } from "../../store";

export default defineComponent({
  props: {
    member: {
      type: Object,
      required: true,
    },
    hasAuthorization: {
      type: Boolean,
      required: true,
    },
  },
  emits: ["update"],
  setup(props, ctx) {
    const showDialog = ref(false);
    const store = useStore();

    const update = () => {
      ctx.emit("update");
      showDialog.value = false;
    };

    const remove = async () => {
      try {
        const tenant = store.getters["auth/tenant"];
        await store.dispatch("namespaces/removeUser", {
          user_id: props.member.id,
          tenant_id: tenant,
        });

        update();
        store.dispatch(
          "snackbar/showSnackbarSuccessAction",
          INotificationsSuccess.namespaceDelete
        );
      } catch (error: any) {
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.namespaceRemoveUser
        );
        throw new Error(error);
      }
    };

    return {
      showDialog,
      remove,
    };
  },
});
</script>
