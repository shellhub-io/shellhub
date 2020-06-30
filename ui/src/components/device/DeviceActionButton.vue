<template>
  <fragment>
    <fragment
      v-if="notificationStatus"
    >
      <v-btn
        x-small
        color="primary"
        @click="dialog = !dialog"
      >
        Accept
      </v-btn>
    </fragment>

    <fragment
      v-else
    >
      <v-btn
        class="mr-2"
        small
        outlined
        @click="dialog = !dialog"
      >
        {{ action }}
      </v-btn>
    </fragment>

    <v-dialog
      v-model="dialog"
      max-width="400"
    >
      <v-card>
        <v-card-title class="headline grey lighten-2 text-center">
          Are you sure?
        </v-card-title>

        <v-card-text class="mt-4 mb-3 pb-1">
          <div>
            You are about to {{ action }} this device
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
      await this.$store.dispatch('devices/accept', this.uid);
      this.refreshDevices();
    },

    async rejectDevice() {
      await this.$store.dispatch('devices/reject', this.uid);
      this.refreshDevices();
    },

    async removeDevice() {
      await this.$store.dispatch('devices/remove', this.uid);
      this.refreshDevices();
    },

    refreshDevices() {
      this.$emit('update');
      if (window.location.pathname === '/devices/pending' || window.location.pathname === '/devices') {
        this.$store.dispatch('devices/refresh');
        this.$store.dispatch('notifications/fetch');
      }

      this.dialog = !this.dialog;
    },
  },
};

</script>
