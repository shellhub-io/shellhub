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
        :disabled="isOwner"
      >
        <template #activator="{ on }">
          <span v-on="on">
            <v-btn
              :disabled="!isOwner"
              class="mr-2"
              small
              outlined
              data-test="tooltipNotOwner-text"
              @click="dialog = !dialog"
            >
              {{ action }}
            </v-btn>
          </span>
        </template>

        <span>
          You are not the owner of this namespace
        </span>
      </v-tooltip>
    </fragment>

    <v-dialog
      v-model="dialog"
      max-width="400"
      data-test="dialog-field"
    >
      <v-card>
        <v-card-title class="headline grey lighten-2 text-center">
          Are you sure?
        </v-card-title>

        <v-card-text class="mt-4 mb-3 pb-1">
          <div data-test="dialog-text">
            You are about to {{ action }} this device.
          </div>
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <v-btn
            text
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

export default {
  name: 'DeviceActionButton',

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
    isOwner() {
      return this.$store.getters['namespaces/owner'];
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
      } catch {
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
