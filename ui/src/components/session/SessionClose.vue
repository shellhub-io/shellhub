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
            Close
          </v-list-item-title>
        </span>

        <span v-on="on">
          <v-icon
            :disabled="!hasAuthorization"
            left
            data-test="close-icon"
            v-on="on"
          >
            mdi-close-circle
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
      <v-card data-test="sessionClose-card">
        <v-card-title class="headline primary">
          Are you sure?
        </v-card-title>

        <v-card-text class="mt-4 mb-3 pb-1">
          You are going to close connection for this device.
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <v-btn
            text
            data-test="cancel-btn"
            @click="close()"
          >
            Cancel
          </v-btn>

          <v-btn
            color="red darken-1"
            text
            data-test="close-btn"
            @click="closeSession()"
          >
            Close
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

import hasPermission from '@/components/filter/permission';

export default {
  name: 'SessionCloseComponent',

  filters: { hasPermission },

  props: {
    uid: {
      type: String,
      required: true,
    },
    device: {
      type: String,
      required: true,
    },

    show: {
      type: Boolean,
      required: true,
    },
  },

  data() {
    return {
      session: {},
      action: 'close',
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
          this.$actions.session[this.action],
        );
      }

      return false;
    },
  },

  created() {
    this.session = {
      uid: this.uid,
      device_uid: this.device,
    };
  },

  methods: {
    async closeSession() {
      try {
        await this.$store.dispatch('sessions/close', this.session);
        this.close();

        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.sessionClose);
        this.$emit('update');
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.sessionClose);
      }
    },

    close() {
      this.$emit('update:show', false);
    },
  },
};

</script>
