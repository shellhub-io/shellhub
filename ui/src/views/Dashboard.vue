<template>
  <fragment>
    <v-alert
      v-if="flag"
      text
      color="#bd4147"
      outlined
      dismissible
    >
      Sorry, we couldn't find the page you were looking for
    </v-alert>
    <v-row
      class="mt-4"
    >
      <v-col
        v-for="(item, index) in items"
        :key="index"
        cols="12"
        md="4"
        class="pt-0"
      >
        <v-card class="pa-2">
          <v-list-item three-line>
            <v-list-item-content>
              <div class="overline mb-4">
                {{ item.title }}
              </div>
              <v-list-item-title class="headline mb-1">
                {{ stats[item.fieldObject] }}
              </v-list-item-title>
              <v-list-item-subtitle>
                {{ item.content }}
              </v-list-item-subtitle>
            </v-list-item-content>

            <v-list-item-avatar
              tile
              size="80"
            >
              <v-icon x-large>
                {{ item.icon }}
              </v-icon>
            </v-list-item-avatar>
          </v-list-item>

          <v-card-actions>
            <div v-if="item.pathName == 'addDevice'">
              <DeviceAdd />
              <v-btn
                class="v-btn--active"
                text
                :data-cy="item.nameUseTest"
                @click="$store.dispatch('modals/showAddDevice', true)"
              >
                {{ item.buttonName }}
              </v-btn>
            </div>

            <div v-else>
              <v-btn
                class="v-btn--active"
                :to="{ name: item.pathName }"
                text
                :data-cy="item.nameUseTest"
              >
                {{ item.buttonName }}
              </v-btn>
            </div>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>

    <Welcome
      v-if="show"
      :dialog="true"
      :curl="curl"
      @finishedEvent="receiveFinish"
    />
  </fragment>
</template>

<script>

import DeviceAdd from '@/components/device/DeviceAdd';
import Welcome from '@/components/welcome/Welcome';

export default {
  name: 'Dashboard',

  components: {
    DeviceAdd,
    Welcome,
  },

  data() {
    return {
      curl: {
        hostname: window.location.hostname,
        tenant: this.$store.getters['auth/tenant'],
      },
      flag: false,
      hasDevicesRegistered: false,
      show: false,
      items: [
        {
          title: 'Registered Devices',
          fieldObject: 'registered_devices',
          content: 'Registered devices into the tenancy account',
          icon: 'devices',
          buttonName: 'Add Device',
          pathName: 'addDevice',
          nameUseTest: 'addDevice-btn',
        },
        {
          title: 'Online Devices',
          fieldObject: 'online_devices',
          content: 'Devices are online and ready for connecting',
          icon: 'devices',
          buttonName: 'View all Devices',
          pathName: 'devices',
          nameUseTest: 'viewDevices-btn',
        },
        {
          title: 'Active Sessions',
          fieldObject: 'active_sessions',
          content: 'Active SSH Sessions opened by users',
          icon: 'devices',
          buttonName: 'View all Sessions',
          pathName: 'sessions',
          nameUseTest: 'viewSessions-btn',
        },
      ],
    };
  },

  computed: {
    stats() {
      return this.$store.getters['stats/stats'];
    },
  },

  async created() {
    try {
      await this.$store.dispatch('stats/get');

      this.hasDevicesRegistered = this.initialState();
      if (localStorage.getItem('onceWelcome') === null) {
        localStorage.setItem('onceWelcome', true);
        this.show = !this.hasDevicesRegistered;
      }
    } catch {
      this.$store.dispatch('modals/showSnackbarErrorLoading', this.$errors.dashboard);
    }
  },

  mounted() {
    this.flag = localStorage.getItem('flag');
    localStorage.removeItem('flag');
  },

  methods: {
    receiveFinish(params) {
      this.hasDevicesRegistered = params;
      this.show = false;
    },

    initialState() {
      return this.stats.registered_devices !== 0
        || this.stats.pending_devices !== 0
        || this.stats.rejected_devices !== 0;
    },
  },
};

</script>
