<template>
  <fragment>
    <v-tooltip
      bottom
    >
      <template #activator="{ on }">
        <span v-on="on">
          <v-icon
            :disabled="!hasAuthorization"
            v-on="on"
            @click="dialog = !dialog"
          >
            delete
          </v-icon>
        </span>
      </template>

      <div>
        <span
          v-if="hasAuthorization"
          data-test="text-tooltip"
        >
          Remove
        </span>

        <span
          v-else
        >
          You don't have this kind of authorization.
        </span>
      </div>
    </v-tooltip>

    <v-dialog
      v-model="dialog"
      max-width="400"
    >
      <v-card data-test="deviceDelete-card">
        <v-card-title class="headline grey lighten-2 text-center">
          Are you sure?
        </v-card-title>

        <v-card-text class="mt-4 mb-3 pb-1">
          You are about to remove this device.
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
  name: 'DeviceDeleteComponent',

  filters: { hasPermission },

  props: {
    uid: {
      type: String,
      required: true,
    },

    redirect: {
      type: Boolean,
    },
  },

  data() {
    return {
      dialog: false,
      action: 'remove',
    };
  },

  computed: {
    hasAuthorization() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.device[this.action],
        );
      }

      return false;
    },
  },

  methods: {
    async remove() {
      try {
        await this.$store.dispatch('devices/remove', this.uid);
        this.dialog = !this.dialog;

        if (this.redirect) {
          this.$router.push('/devices');
        }

        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.deviceDelete);
        this.$emit('update');
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.deviceDelete);
      }
    },
  },
};

</script>
