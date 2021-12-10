<template>
  <v-dialog
    v-if="hasAuthorization"
    v-model="show"
    persistent
    max-width="900"
  >
    <v-card data-test="deviceChooser-dialog">
      <v-card-title class="headline grey lighten-2 text-center">
        Update account or select three devices
      </v-card-title>

      <v-card-text class="mt-4 pb-1">
        <p>
          You currently have no subscription to the
          <a :href="url()">
            premium plan
          </a> and the free version is limited to 3 devices. To unlock access to all
          devices, you can subscribe to the
          <a :href="url()">
            premium plan
          </a>. Case, If you want to continue on the free plan, you need to select three devices.
        </p>
      </v-card-text>

      <v-app-bar
        flat
        color="transparent"
        class="mt-0"
      >
        <v-tabs
          centered
        >
          <v-tab
            v-for="item in items"
            :key="item.title"
            :data-test="item.title+'-tab'"
            @click="doAction(item.action)"
          >
            {{ item.title }}
          </v-tab>
        </v-tabs>
      </v-app-bar>

      <v-card-text class="mb-2 pb-0">
        <DeviceListChooser
          :action="action"
          data-test="deviceListChooser-component"
        />
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

        <v-tooltip
          :disabled="!disableTooltipOrButton"
          top
        >
          <template #activator="{ on, attrs }">
            <span v-on="on">
              <v-btn
                text
                v-bind="attrs"
                :disabled="disableTooltipOrButton"
                data-test="accept-btn"
                v-on="on"
                @click="accept()"
              >
                Accept
              </v-btn>
            </span>
          </template>

          <span>
            You can select 3 devices or less.
          </span>
        </v-tooltip>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>

import DeviceListChooser from '@/components/device/DeviceListChooser';

import hasPermission from '@/components/filter/permission';

export default {
  name: 'DeviceChooserComponent',

  filters: { hasPermission },

  components: {
    DeviceListChooser,
  },

  data() {
    return {
      hostname: window.location.hostname,
      action: 'suggestedDevices',
      dialog: false,
      items: [
        {
          title: 'Suggested Devices',
          action: 'suggestedDevices',
        },
        {
          title: 'All devices',
          action: 'allDevices',
        },
      ],
      permissionAction: 'chooser',
    };
  },

  computed: {
    show: {
      get() {
        return this.$store.getters['devices/getDeviceChooserStatus'];
      },

      set(value) {
        this.$store.dispatch('devices/setDeviceChooserStatus', value);
      },
    },

    disableTooltipOrButton() {
      return (this.$store.getters['devices/getDevicesSelected'].length <= 0
        || this.$store.getters['devices/getDevicesSelected'].length > 3)
        && this.action !== this.items[0].action;
    },

    equalThreeDevices() {
      return this.$store.getters['devices/getDevicesSelected'].length === 3;
    },

    hasAuthorization() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.device[this.permissionAction],
        );
      }

      return false;
    },
  },

  watch: {
    show(status) {
      if (status && this.action === this.items[0].action) {
        this.$store.dispatch('devices/getDevicesMostUsed');
      }
    },
  },

  methods: {
    async doAction(action) {
      this.action = action;

      if (action === this.items[0].action) {
        this.$store.dispatch('devices/getDevicesMostUsed');
      } else if (action === this.items[1].action) {
        const data = {
          perPage: 10,
          page: 1,
          filter: this.$store.getters['devices/getFilter'],
          status: 'accepted',
          sortStatusField: null,
          sortStatusString: 'asc',
        };

        try {
          await this.$store.dispatch('devices/setDevicesForUserToChoose', data);
        } catch (error) {
          if (error.response.status === 403) {
            this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
          } else {
            this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.deviceList);
          }
        }
      } else {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.deviceList);
      }
    },

    accept() {
      if (this.action === this.items[0].action) {
        this.sendDevicesChoice(this.$store.getters['devices/getDevicesForUserToChoose']);
      } else {
        this.sendDevicesChoice(this.$store.getters['devices/getDevicesSelected']);
      }
    },

    async sendDevicesChoice(devices) {
      const choices = [];
      devices.forEach((device) => {
        choices.push(device.uid);
      });

      try {
        await this.$store.dispatch('devices/postDevicesChooser', { choices });
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.deviceChoose);

        this.$store.dispatch('devices/refresh');

        this.close();
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.deviceChooser);
      }

      this.$store.dispatch('stats/get');
    },

    url() {
      return `${window.location.protocol}//${this.hostname}/settings/billing`;
    },

    close() {
      this.$store.dispatch('devices/setDeviceChooserStatus', false);
    },
  },
};

</script>
