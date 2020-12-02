<template>
  <fragment>
    <v-tooltip bottom>
      <template #activator="{ on }">
        <v-icon
          v-on="on"
          @click="dialog = !dialog"
        >
          delete
        </v-icon>
      </template>
      <span>Remove</span>
    </v-tooltip>

    <v-dialog
      v-model="dialog"
      max-width="400"
    >
      <v-card>
        <v-card-title class="headline grey lighten-2 text-center">
          Are you sure?
        </v-card-title>

        <v-card-text class="mt-4 mb-3 pb-1">
          You are about to remove this {{ action }} key
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <v-btn
            text
            @click="dialog=!dialog"
          >
            Close
          </v-btn>

          <v-btn
            color="red darken-1"
            text
            @click="remove();"
          >
            Remove
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

export default {
  name: 'PublicKeyDelete',

  props: {
    fingerprint: {
      type: String,
      required: true,
    },

    action: {
      type: String,
      default: 'public',
      required: false,
      validator: (value) => ['public', 'private'].includes(value),
    },
  },

  data() {
    return {
      dialog: false,
    };
  },

  methods: {
    async remove() {
      switch (this.action) {
      case 'public':
        try {
          await this.$store.dispatch('publickeys/remove', this.fingerprint);
          this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.publicKeyDeleting);
          this.dialog = false;
        } catch {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.publicKeyDeleting);
        }
        break;
      case 'private':
        try {
          await this.$store.dispatch('privatekeys/remove', this.id);
          this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.privateKeyDeleting);
          this.dialog = false;
        } catch {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.privateKeyDeleting);
        }
        break;
      default:
      }
    },
  },
};
</script>
