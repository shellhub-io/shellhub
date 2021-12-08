<template>
  <fragment>
    <v-alert
      v-if="notFound"
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
                data-test="addDevice-btn"
              />
            </div>

            <div v-else>
              <v-btn
                class="v-btn--active"
                :to="{ name: item.pathName }"
                text
                color="primary"
                small
                :data-test="item.nameUseTest"
              >
                {{ item.buttonName }}
              </v-btn>
            </div>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>
  </fragment>
</template>

<script>

import DeviceAdd from '@/components/device/DeviceAdd';

export default {
  name: 'DashboardView',

  components: {
    DeviceAdd,
  },

  props: {
    notFound: {
      type: Boolean,
      required: false,
      default: false,
    },
  },

  data() {
    return {
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

    currentInANamespace() {
      return localStorage.getItem('tenant') !== '';
    },

    hasNamespace() {
      return this.$store.getters['namespaces/getNumberNamespaces'] > 0;
    },
  },

  watch: {
    hasNamespace(status) {
      if (status) {
        this.$store.dispatch('stats/get');
      }
    },
  },

  created() {
    this.$store.dispatch('users/setStatusUpdateAccountDialog', true);
  },

  mounted() {
    this.flag = localStorage.getItem('flag');
    localStorage.removeItem('flag');
  },
};

</script>
