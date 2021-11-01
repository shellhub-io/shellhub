<template>
  <fragment>
    <v-card
      v-if="showMessage"
      color="transparent"
      class="elevation-0"
      data-test="accountCreated-card"
    >
      <v-card-title class="justify-center">
        Account Creation Successful
      </v-card-title>

      <v-card-text class="d-flex align-center justify-center">
        Thank you for registaring an account on ShellHub.
        An email was sent with a confirmation link. You need to click on the link
        to activate your account.
      </v-card-text>

      <v-card-text class="d-flex align-center justify-center">
        If you haven't received the email, click on the button.
      </v-card-text>

      <v-card-actions class="justify-center">
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
  </fragment>
</template>

<script>

export default {
  name: 'AccountCreatedComponent',

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

  computed: {
    showMessage: {
      get() {
        return this.show;
      },

      set(value) {
        this.$emit('show', value);
      },
    },
  },

  methods: {
    async resendEmail() {
      try {
        await this.$store.dispatch('users/resendEmail', this.username);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.resendEmail);
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorDefault');
      }
    },
  },
};

</script>
