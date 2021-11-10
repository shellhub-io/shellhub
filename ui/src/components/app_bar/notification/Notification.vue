<template>
  <v-menu
    offset-y
    :close-on-content-click="false"
    :value="shown"
  >
    <template #activator="{ on }">
      <v-badge
        :content="showNumberNotifications"
        :value="showNumberNotifications"
        overlap
        offset-x="20"
        color="success"
        data-test="notifications-badge"
      >
        <v-chip
          v-on="on"
          @click="getNotifications()"
        >
          <v-icon
            class="ml-2 mr-2"
            :size="defaultSize"
          >
            notifications
          </v-icon>
        </v-chip>
      </v-badge>
    </template>

    <v-card
      v-if="!getStatusNotifications"
      data-test="hasNotifications-subheader"
    >
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
                <router-link
                  :to="{ name: 'detailsDevice', params: { id: item.uid } }"
                  :data-test="item.uid+'-field'"
                >
                  {{ item.name }}
                </router-link>
              </v-list-item-title>
            </v-list-item-content>

            <v-list-item-action>
              <DeviceActionButton
                v-if="hasAuthorization"
                :uid="item.uid"
                :notification-status="true"
                action="accept"
                :data-test="item.uid+'-btn'"
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
        data-test="show-btn"
        @click="shown=false"
      >
        Show all Pending Devices
      </v-btn>
    </v-card>

    <v-card
      v-else
      data-test="noNotifications-subheader"
    >
      <v-subheader>
        You don't have notifications
      </v-subheader>
    </v-card>
  </v-menu>
</template>

<script>

import DeviceActionButton from '@/components/device/DeviceActionButton';

import hasPermission from '@/components/filter/permission';

export default {
  name: 'NotificationIcon',

  filters: { hasPermission },

  components: {
    DeviceActionButton,
  },

  data() {
    return {
      listNotifications: [],
      numberNotifications: 0,
      shown: false,
      inANamespace: false,
      defaultSize: 24,
    };
  },

  computed: {
    getListNotifications() {
      return this.$store.getters['notifications/list'];
    },

    getNumberNotifications() {
      return this.$store.getters['notifications/getNumberNotifications'];
    },

    showNumberNotifications() {
      const numberNotifications = this.getNumberNotifications;
      const pendingDevices = this.$store.getters['stats/stats'].pending_devices;

      if (numberNotifications === 0 && pendingDevices !== undefined) {
        return this.$store.getters['stats/stats'].pending_devices;
      }

      return numberNotifications;
    },

    getStatusNotifications() {
      if (this.getNumberNotifications === 0) { return true; }
      return false;
    },

    hasNamespace() {
      return this.$store.getters['namespaces/getNumberNamespaces'] !== 0;
    },

    hasAuthorization() {
      const accessType = this.$store.getters['auth/accessType'];
      if (accessType !== '') {
        return hasPermission(
          this.$authorizer.accessType[accessType],
          this.$actions.notification.view,
        );
      }

      return false;
    },
  },

  watch: {
    hasNamespace(status) {
      this.inANamespace = status;
    },
  },

  methods: {
    async getNotifications() {
      if (this.hasNamespace) {
        try {
          await this.$store.dispatch('notifications/fetch');
          this.shown = true;
        } catch (error) {
          switch (true) {
          case (!this.inANamespace && error.response.status === 403): { // dialog pops
            break;
          }
          case (error.response.status === 403): {
            this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
            break;
          }
          default: {
            this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.notificationList);
          }
          }
        }
      }
    },

    refresh() {
      if (this.hasNamespace) {
        this.getNotifications();

        if (this.getNumberNotifications === 0) {
          this.$store.dispatch('stats/get');
        }
      }
    },
  },
};

</script>
