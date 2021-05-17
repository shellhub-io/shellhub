<template>
  <v-menu
    offset-y
    :close-on-content-click="false"
    :value="shown"
  >
    <template #activator="{ on }">
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

    <v-card
      v-if="!getStatusNotifications"
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
                v-if="isOwner"
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
        @click="shown=false"
      >
        Show all Pending Devices
      </v-btn>
    </v-card>
    <v-card
      v-else
    >
      <v-subheader
        data-test="noNotifications"
      >
        You don't have notifications
      </v-subheader>
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
      inANamespace: false,
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

    isOwner() {
      return this.$store.getters['namespaces/owner'];
    },

    hasNamespace() {
      return this.$store.getters['namespaces/getNumberNamespaces'] !== 0;
    },
  },

  watch: {
    hasNamespace(status) {
      this.inANamespace = status;

      if (status) {
        this.getNotifications();
      }
    },
  },

  methods: {
    async getNotifications() {
      try {
        await this.$store.dispatch('notifications/fetch');
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
    },

    refresh() {
      this.getNotifications();
    },
  },
};

</script>
