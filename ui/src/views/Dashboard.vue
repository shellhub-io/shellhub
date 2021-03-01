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
        <v-card
          outlined
          elevation="1"
          :disabled="!currentInANamespace"
        >
          <v-list-item three-line>
            <v-list-item-content>
              <div class="overline mb-4">
                {{ item.title }}
              </div>
              <v-list-item-title class="headline mb-1">
                {{ stats[item.fieldObject] || 0 }}
              </v-list-item-title>
              <v-list-item-subtitle class="grey--text">
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

          <v-card-actions class="ma-2">
            <div v-if="item.pathName == 'addDevice'">
              <DeviceAdd
                :small-button="true"
                data-cy="addDevice-btn"
              />
            </div>

            <div v-else>
              <v-btn
                class="v-btn--active"
                :to="{ name: item.pathName }"
                text
                color="primary"
                small
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
      :show.sync="show"
    />
    <NamespaceInstructions
      :show.sync="showInstructions"
    />
  </fragment>
</template>

<script>

import DeviceAdd from '@/components/device/DeviceAdd';
import Welcome from '@/components/welcome/Welcome';
import NamespaceInstructions from '@/components/app_bar/namespace/NamespaceInstructions';

export default {
  name: 'Dashboard',

  components: {
    DeviceAdd,
    Welcome,
    NamespaceInstructions,
  },

  data() {
    return {
      flag: false,
      show: false,
      showInstructions: false,
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

    hasNamespaces() {
      return this.$store.getters['namespaces/getNumberNamespaces'] !== 0;
    },

    currentInANamespace() {
      return localStorage.getItem('tenant') !== '';
    },

    onceShow() {
      return localStorage.getItem('noNamespace') !== null;
    },
  },

  async created() {
    try {
      await this.$store.dispatch('stats/get');
      this.showScreenWelcome();
    } catch (e) {
      if (this.onceShow) {
        switch (true) {
        case (e.response.status === 403): {
          this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
          break;
        }
        default: {
          this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.dashboard);
        }
        }
      } else {
        this.showNamespaceInstructions();
      }
    }
  },

  mounted() {
    this.flag = localStorage.getItem('flag');
    localStorage.removeItem('flag');
  },

  methods: {
    namespaceHasBeenShown(tenant) {
      return JSON.parse(localStorage.getItem('namespacesWelcome'))[tenant] !== undefined;
    },

    showNamespaceInstructions() {
      let status = false;

      if (localStorage.getItem('noNamespace') === null && !this.hasNamespaces) {
        localStorage.setItem('noNamespace', true);

        status = true;
      }
      this.showInstructions = status;
    },

    hasDevices() {
      return this.stats.registered_devices !== 0
        || this.stats.pending_devices !== 0
        || this.stats.rejected_devices !== 0;
    },

    async getNamespace() {
      try {
        await this.$store.dispatch('namespaces/get', localStorage.getItem('tenant'));
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.namespaceLoad);
      }
    },

    async showScreenWelcome() {
      let status = false;

      await this.getNamespace();
      const tenantID = await this.$store.getters['namespaces/get'].tenant_id;

      if (!this.namespaceHasBeenShown(tenantID) && !this.hasDevices()) {
        this.$store.dispatch('auth/setShowWelcomeScreen', tenantID);
        status = true;
      }

      this.show = status;
    },
  },
};

</script>
