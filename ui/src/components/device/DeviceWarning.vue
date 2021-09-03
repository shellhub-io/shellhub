<template>
  <fragment>
    <v-dialog
      v-model="show"
      max-width="900"
    >
      <v-card data-test="deviceWarning-dialog">
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
          <DeviceListChoice
            :action="action"
            data-test="deviceListChoice-component"
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
            :disabled="!showTooltip"
            top
          >
            <template #activator="{ on, attrs }">
              <v-btn
                text
                v-bind="attrs"
                data-test="accept-btn"
                v-on="on"
                @click="accept()"
              >
                Accept
              </v-btn>
            </template>

            <span>
              You need to select 3 devices.
            </span>
          </v-tooltip>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

import DeviceListChoice from '@/components/device/DeviceListChoice';

export default {
  name: 'DeviceWarning',

  components: {
    DeviceListChoice,
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
    };
  },

  computed: {
    show: {
      get() {
        return this.$store.getters['devices/getDeviceWarning'];
      },

      set(value) {
        this.$store.dispatch('devices/setDeviceWarning', value);
      },
    },

    showTooltip() {
      return this.$store.getters['devices/getDevicesSelected'].length !== 3
        && this.action !== this.items[0].action;
    },

    equalThreeDevices() {
      return this.$store.getters['devices/getDevicesSelected'].length === 3;
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
          await this.$store.dispatch('devices/fetch', data);
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
        this.sendDevicesChoice(this.$store.getters['devices/list']);
      } else if (this.$store.getters['devices/getDevicesSelected'].length === 3) {
        this.sendDevicesChoice(this.$store.getters['devices/getDevicesSelected']);
      } else {
        this.$store.dispatch('snackbar/showSnackbarDeviceChoice');
      }
    },

    async sendDevicesChoice(devices) {
      const choices = [];
      devices.forEach((device) => {
        choices.push(device.uid);
      });

      try {
        await this.$store.dispatch('devices/postDevicesChoice', { choices });
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.deviceChoice);

        this.close();
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.deviceChoice);
      }

      this.$store.dispatch('stats/get');
    },

    url() {
      return `${window.location.protocol}//${this.hostname}/settings/billing`;
    },

    close() {
      this.$store.dispatch('devices/setDeviceWarning', false);
    },
  },
};

</script>
