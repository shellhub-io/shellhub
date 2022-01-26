<template>
  <fragment>
    <v-tooltip
      :disabled="hasAuthorization"
      bottom
    >
      <template #activator="{ on }">
        <span v-on="on">
          <v-list-item-title
            data-test="close-item"
            v-on="on"
          >
            Remove
          </v-list-item-title>
        </span>

        <span v-on="on">
          <v-icon
            :disabled="!hasAuthorization"
            left
            data-test="remove-icon"
            v-on="on"
          >
            delete
          </v-icon>
        </span>
      </template>

      <span v-if="!hasAuthorization && action == 'public'">
        You don't have this kind of authorization.
      </span>
    </v-tooltip>

    <v-dialog
      v-model="showDialog"
      max-width="400"
      @click:outside="close"
    >
      <v-card data-test="keyDelete-card">
        <v-card-title class="headline primary">
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
            @click="close()"
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

    show: {
      type: Boolean,
      required: true,
    },
  },

  computed: {
    showDialog: {
      get() {
        return this.show && this.hasAuthorization;
      },
      set(value) {
        this.$emit('update:show', value);
      },
    },

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
      this.$emit('update:show', false);
    },
  },
};
</script>
