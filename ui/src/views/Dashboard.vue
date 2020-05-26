<template>
  <fragment>
    <v-alert
      v-if="flag"
      type="warning"
      outlined
    >
      Sorry, we couldn't find the page you were looking for
    </v-alert>
    <v-row>
      <v-col
        cols="12"
        md="4"
        class="pt-0"
      >
        <v-card class="pa-2">
          <v-list-item three-line>
            <v-list-item-content>
              <div class="overline mb-4">
                Registered Devices
              </div>
              <v-list-item-title class="headline mb-1">
                {{ stats.registered_devices }}
              </v-list-item-title>
              <v-list-item-subtitle>Registered devices into the tenancy account</v-list-item-subtitle>
            </v-list-item-content>

            <v-list-item-avatar
              tile
              size="80"
            >
              <v-icon x-large>
                devices
              </v-icon>
            </v-list-item-avatar>
          </v-list-item>

          <v-card-actions>
            <DeviceAdd />
            <v-btn
              text
              data-cy="addDevice-btn"
              @click="$store.dispatch('modals/showAddDevice', true)"
            >
              Add Device
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-col>

      <v-col
        cols="12"
        md="4"
        class="pt-0"
      >
        <v-card class="pa-2">
          <v-list-item three-line>
            <v-list-item-content>
              <div class="overline mb-4">
                Online Devices
              </div>
              <v-list-item-title class="headline mb-1">
                {{ stats.online_devices }}
              </v-list-item-title>
              <v-list-item-subtitle>Devices are online and ready for connecting</v-list-item-subtitle>
            </v-list-item-content>

            <v-list-item-avatar
              tile
              size="80"
            >
              <v-icon x-large>
                devices
              </v-icon>
            </v-list-item-avatar>
          </v-list-item>

          <v-card-actions>
            <v-btn
              to="/devices"
              text
              data-cy="viewDevices-btn"
            >
              View all Devices
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-col>

      <v-col
        cols="12"
        md="4"
        class="pt-0"
      >
        <v-card class="pa-2">
          <v-list-item three-line>
            <v-list-item-content>
              <div class="overline mb-4">
                Active Sessions
              </div>
              <v-list-item-title class="headline mb-1">
                {{ stats.active_sessions }}
              </v-list-item-title>
              <v-list-item-subtitle>Active SSH Sessions opened by users</v-list-item-subtitle>
            </v-list-item-content>

            <v-list-item-avatar
              tile
              size="80"
            >
              <v-icon x-large>
                history
              </v-icon>
            </v-list-item-avatar>
          </v-list-item>

          <v-card-actions>
            <v-btn
              to="/sessions"
              text
              data-cy="viewSessions-btn"
            >
              View all Sessions
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>
  </fragment>
</template>

<script>
import DeviceAdd from '@/components/device/DeviceAdd.vue';

export default {
  name: 'Dashboard',

  components: {
    DeviceAdd
  },

  data(){
    return {
      flag: false
    };
  },

  computed: {
    stats() {
      return this.$store.getters['stats/stats'];
    }
  },

  created() {
    this.$store.dispatch('stats/get');
  },

  mounted() {
    this.flag = localStorage.getItem('flag');
    localStorage.removeItem('flag');
  }
};
</script>

