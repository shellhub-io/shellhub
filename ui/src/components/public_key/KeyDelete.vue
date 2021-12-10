<template>
  <fragment>
    <v-tooltip
      bottom
    >
      <template #activator="{ on }">
        <span v-on="on">
          <v-icon
            :disabled="!hasAuthorization && action == 'public'"
            v-on="on"
            @click="dialog = !dialog"
          >
            delete
          </v-icon>
        </span>
      </template>

      <div>
        <span
          v-if="hasAuthorization || action == 'private'"
          data-test="text-tooltip"
        >
          Remove
        </span>

        <span v-else>
          You don't have this kind of authorization.
        </span>
      </div>
    </v-tooltip>

    <v-dialog
      v-model="dialog"
      max-width="400"
    >
      <v-card data-test="keyDelete-card">
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
            data-test="close-btn"
            @click="dialog=!dialog"
          >
            Close
          </v-btn>

          <v-btn
            color="red darken-1"
            text
            data-test="remove-btn"
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

import hasPermission from '@/components/filter/permission';

export default {
  name: 'KeyDeleteComponent',

  filters: { hasPermission },

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
    hasAuthorization() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.publicKey.remove,
        );
      }

      return false;
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
