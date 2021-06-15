<template>
  <fragment>
    <v-tooltip
      bottom
    >
      <template #activator="{ on }">
        <span v-on="on">
          <v-icon
            :disabled="!isOwner && action == 'public'"
            v-on="on"
            @click="dialog = !dialog"
          >
            delete
          </v-icon>
        </span>
      </template>

      <div>
        <span v-if="isOwner || action == 'private'">
          Remove
        </span>

        <span v-else>
          You are not the owner of this namespace
        </span>
      </div>
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
          You are about to remove this {{ action }} key.
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
  name: 'KeyDelete',

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

  computed: {
    isOwner() {
      return this.$store.getters['namespaces/owner'];
    },
  },

  methods: {
    async remove() {
      switch (this.action) {
      case 'public':
        try {
          await this.$store.dispatch('publickeys/remove', this.fingerprint);
          this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.publicKeyDeleting);
          this.update();
        } catch {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.publicKeyDeleting);
        }
        break;
      case 'private':
        try {
          await this.$store.dispatch('privatekeys/remove', this.fingerprint);
          this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.privateKeyDeleting);
          this.close();
        } catch {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.privateKeyDeleting);
        }
        break;
      default:
      }
    },

    update() {
      this.$emit('update');
      this.close();
    },

    close() {
      this.dialog = !this.dialog;
    },
  },
};
</script>
