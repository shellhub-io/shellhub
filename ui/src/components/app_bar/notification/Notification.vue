<template>
  <v-menu
    offset-y
    :close-on-content-click="false"
    :value="shown"
    :disabled="getStatusNotifications"
  >
    <template v-slot:activator="{ on }">
      <v-badge
        :content="getNumberNotifications"
        :value="getNumberNotifications"
        overlap
        offset-x="20"
        color="success"
      >
        <v-chip
          v-on="on"
          @click="shown=true"
        >
          <v-icon>
            notifications
          </v-icon>
        </v-chip>
      </v-badge>
    </template>

    <v-card>
      <v-subheader>Pending Devices</v-subheader>

      <v-divider />

      <v-list class="pa-0">
        <v-list-item-group :v-model="1">
          <v-list-item
            v-for="item in getListNotifications"
            :key="item.uid"
          >
            <v-list-item-content>
              <v-list-item-title>
                <router-link :to="{ name: 'detailsDevice', params: { id: item.uid } }">
                {{ item.name }}
                </router-link>
              </v-list-item-title>
            </v-list-item-content>

            <v-list-item-action>
              <DeviceActionButton
                :uid="item.uid"
                :notification-status="true"
                action="accept"
                @update="refresh"
              />
            </v-list-item-action>
          </v-list-item>
        </v-list-item-group>
      </v-list>

      <v-divider />

      <v-btn
        to="/devices/pending"
        block
        link
        small
        @click="shown=false"
      >
        Show all Pending Devices
      </v-btn>
    </v-card>
  </v-menu>
</template>

<script>

import DeviceActionButton from '@/components/device/DeviceActionButton';

export default {
  name: 'NotificationIcon',

  components: {
    DeviceActionButton,
  },

  data() {
    return {
      listNotifications: [],
      numberNotifications: 0,
      shown: false,
    };
  },

  computed: {
    getListNotifications() {
      return this.$store.getters['notifications/list'];
    },

    getNumberNotifications() {
      return this.$store.getters['notifications/getNumberNotifications'];
    },

    getStatusNotifications() {
      if (this.getNumberNotifications === 0) { return true; }
      return false;
    },
  },

  async created() {
    await this.getNotifications();
  },

  methods: {
    getNotifications() {
      this.$store.dispatch('notifications/fetch');
    },

    refresh() {
      this.getNotifications();
    },
  },
};

</script>
