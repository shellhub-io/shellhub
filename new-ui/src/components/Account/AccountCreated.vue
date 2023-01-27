<template>
  <v-card
    v-if="showMessage"
    color="transparent"
    class="elevation-0 bg-v-theme-surface mt-2"
    data-test="accountCreated-card"
  >
    <v-card-title class="text-h5 pa-3 bg-primary">
      Account Creation Successful
    </v-card-title>

    <v-card-text class="mt-4 mb-0 pb-1">
      <p class="text-caption mt-2 mb-0">
        Thank you for registering an account on ShellHub. An email was sent with
        a confirmation link. You need to click on the link to activate your
        account.
      </p>

      <p class="text-caption mt-2 mb-0">
        If you haven't received the email, click on the button.
      </p>
    </v-card-text>

    <v-card-actions class="justify-center">
      <v-spacer />
      <v-btn
        type="submit"
        color="primary"
        data-test="resendEmail-btn"
        @click="resendEmail()"
      >
        RESEND EMAIL
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts">
import { defineComponent, computed } from "vue";
import { INotificationsSuccess } from "../../interfaces/INotifications";
import { useStore } from "../../store";

export default defineComponent({
  props: {
    show: {
      type: Boolean,
      required: true,
    },
    username: {
      type: String,
      required: true,
    },
  },
  emits: ["show"],
  setup(props, { emit }) {
    const store = useStore();

    const showMessage = computed({
      get: () => props.show,
      set: (value: boolean) => emit("show", value),
    });

    const resendEmail = async () => {
      try {
        await store.dispatch("users/resendEmail", props.username);
        store.dispatch(
          "snackbar/showSnackbarSuccessAction",
          INotificationsSuccess.resendEmail,
        );
      } catch (error: any) {
        store.dispatch("snackbar/showSnackbarErrorDefault");
        throw new Error(error);
      }
    };

    return {
      resendEmail,
      showMessage,
    };
  },
});
</script>
