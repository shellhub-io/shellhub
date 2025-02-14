<template>
  <v-tooltip bottom anchor="bottom">
    <template v-slot:activator="{ props }">
      <v-icon tag="a" dark v-bind="props" @click="dialog = !dialog">mdi-delete </v-icon>
    </template>
    <span>Remove</span>
  </v-tooltip>

  <v-dialog max-width="450" v-model="dialog">
    <v-card>
      <v-card-title class="lighten-2 text-center mt-2"> Are you sure? </v-card-title>
      <v-divider />

      <v-card-text class="mt-4 mb-3 pb-1">
        You are about to remove this user and all its namespace associated data. This action will
        also bill usage for namespace subscriptions if any and might take some time to finish.
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn text @click="dialog = !dialog"> Close </v-btn>

        <v-btn color="red darken-1" text @click="remove()"> Remove </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { defineComponent, ref } from "vue";
import { useRouter } from "vue-router";
import { INotificationsError, INotificationsSuccess } from "../../interfaces/INotifications";
import { useStore } from "../../store";

export default defineComponent({
  props: {
    id: {
      type: String,
      required: true,
    },

    redirect: {
      type: Boolean,
    },
  },
  emits: ["update"],
  setup(props, ctx) {
    const dialog = ref(false);
    const store = useStore();
    const router = useRouter();

    const remove = async () => {
      dialog.value = !dialog.value;

      try {
        await store.dispatch("users/remove", props.id);

        if (props.redirect) {
          router.push("/users");
        }

        store.dispatch("snackbar/showSnackbarSuccessAction", INotificationsSuccess.userDelete);
        await store.dispatch("users/refresh");
        ctx.emit("update");
      } catch {
        store.dispatch("snackbar/showSnackbarErrorAction", INotificationsError.userDelete);
      }
    };

    return {
      dialog,
      remove,
    };
  },
});
</script>
