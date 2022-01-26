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

      <span v-if="!hasAuthorization">
        You don't have this kind of authorization.
      </span>
    </v-tooltip>

    <v-dialog
      v-model="showDialog"
      max-width="400"
      @click:outside="close"
    >
      <v-card data-test="deviceDelete-card">
        <v-card-title class="headline primary">
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

    show: {
      type: Boolean,
      required: true,
    },
  },

  data() {
    return {
      action: 'remove',
    };
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
        this.close();

        if (this.redirect) {
          this.$router.push('/devices');
        }

        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.deviceDelete);
        this.$emit('update');
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.deviceDelete);
      }
    },

    close() {
      this.$emit('update:show', false);
    },
  },
};

</script>
