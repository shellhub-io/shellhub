<template>
  <fragment>
    <fragment
      v-if="notificationStatus"
    >
      <v-btn
        x-small
        color="primary"
        data-test="notification-btn"
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
            <v-list-item-title
              data-test="action-item"
              v-on="on"
            >
              {{ action | capitalizeFirstLetter }}
            </v-list-item-title>
          </span>

          <span v-on="on">
            <v-icon
              :disabled="!hasAuthorization"
              left
              data-test="action-icon"
              v-on="on"
            >
              {{ icon }}
            </v-icon>
          </span>
        </template>

        <span v-if="!hasAuthorization">
          You don't have this kind of authorization.
        </span>
      </v-tooltip>
    </fragment>

    <v-dialog
      v-model="showDialog"
      max-width="400"
      @click:outside="close"
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
            @click="close()"
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

import capitalizeFirstLetter from '@/components/filter/string';
import hasPermission from '@/components/filter/permission';

export default {
  name: 'DeviceActionButtonComponent',

  filters: { capitalizeFirstLetter, hasPermission },

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

    show: {
      type: Boolean,
      required: true,
    },
  },

  data() {
    return {
      icon: this.findIcon(),
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
        this.close();

        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.deviceAccepting);
      }
    },

    async rejectDevice() {
      try {
        await this.$store.dispatch('devices/reject', this.uid);
        this.refreshStats();
        this.refreshDevices();
      } catch {
        this.close();

        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.deviceRejecting);
      }
    },

    async removeDevice() {
      try {
        await this.$store.dispatch('devices/remove', this.uid);
        this.refreshDevices();
      } catch {
        this.close();

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

        this.close();
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

    findIcon() {
      switch (this.action) {
      case 'accept':
        return 'mdi-check';
      case 'reject':
        return 'close';
      case 'remove':
        return 'delete';
      default:
        return '';
      }
    },

    close() {
      this.$emit('update:show', false);
    },
  },
};

</script>
