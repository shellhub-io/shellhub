<template>
  <fragment>
    <fragment
      v-if="notificationStatus"
    >
      <v-btn
        x-small
        color="primary"
        data-test="notification-btn"
        @click="dialog = !dialog"
      >
        Accept
      </v-btn>
    </fragment>

    <fragment
      v-else
    >
      <v-tooltip
        bottom
        :disabled="hasAuthorization"
      >
        <template #activator="{ on }">
          <span v-on="on">
            <v-btn
              :disabled="!hasAuthorization"
              class="mr-2"
              small
              outlined
              data-test="tooltip-text"
              @click="dialog = !dialog"
            >
              {{ action }}
            </v-btn>
          </span>
        </template>

        <span>
          You don't have this kind of authorization.
        </span>
      </v-tooltip>
    </fragment>

    <v-dialog
      v-model="dialog"
      max-width="400"
    >
      <v-card data-test="deviceActionButton-card">
        <v-card-title class="headline grey lighten-2 text-center">
          Are you sure?
        </v-card-title>

        <v-card-text class="mt-4 mb-3 pb-1">
          <div>
            You are about to {{ action }} this device.
          </div>
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <v-btn
            text
            data-test="cancel-btn"
            @click="dialog=!dialog"
          >
            Cancel
          </v-btn>

          <v-btn
            text
            data-test="dialog-btn"
            @click="doAction();"
          >
            {{ action }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

import hasPermission from '@/components/filter/permission';

export default {
  name: 'DeviceActionButtonComponent',

  filters: { hasPermission },

  props: {
    uid: {
      type: String,
      required: true,
    },

    notificationStatus: {
      type: Boolean,
      required: false,
      default: false,
    },

    action: {
      type: String,
      default: 'accept',
      validator: (value) => ['accept', 'reject', 'remove'].includes(value),
    },
  },

  data() {
    return {
      dialog: false,
    };
  },

  computed: {
    hasAuthorization() {
      const accessType = this.$store.getters['auth/accessType'];
      if (accessType !== '') {
        return hasPermission(
          this.$authorizer.accessType[accessType],
          this.$actions.device[this.action],
        );
      }

      return false;
    },
  },

  methods: {
    doAction() {
      switch (this.action) {
      case 'accept':
        this.acceptDevice();
        break;
      case 'reject':
        this.rejectDevice();
        break;
      case 'remove':
        this.removeDevice();
        break;
      default:
      }
    },

    async acceptDevice() {
      try {
        await this.$store.dispatch('devices/accept', this.uid);
        this.refreshStats();
        this.refreshDevices();
      } catch (error) {
        if (error.response.status === 402) {
          this.$store.dispatch('users/setStatusUpdateAccountDialogByDeviceAction', true);
        }

        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.deviceAccepting);
      }
    },

    async rejectDevice() {
      try {
        await this.$store.dispatch('devices/reject', this.uid);
        this.refreshStats();
        this.refreshDevices();
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.deviceRejecting);
      }
    },

    async removeDevice() {
      try {
        await this.$store.dispatch('devices/remove', this.uid);
        this.refreshDevices();
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.deviceDelete);
      }
    },

    refreshDevices() {
      try {
        this.$emit('update');
        if (window.location.pathname === '/devices/pending' || window.location.pathname === '/devices') {
          this.$store.dispatch('devices/refresh');
          this.$store.dispatch('notifications/fetch');
        }

        this.dialog = !this.dialog;
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.deviceList);
      }
    },

    async refreshStats() {
      try {
        await this.$store.dispatch('stats/get');
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorDefault');
      }
    },
  },
};

</script>
