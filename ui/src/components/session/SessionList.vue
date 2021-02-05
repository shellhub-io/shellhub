<template>
  <fragment>
    <div class="d-flex pa-0 align-center">
      <h1>Sessions</h1>
      <v-spacer />
      <v-spacer />
    </div>

    <v-card class="mt-2">
      <v-app-bar
        flat
        color="transparent"
      >
        <v-toolbar-title />
      </v-app-bar>

      <v-divider />

      <v-card-text class="pa-0">
        <v-data-table
          :headers="headers"
          :items="getListSessions"
          data-test="dataTable-field"
          item-key="uid"
          :sort-by="['started_at']"
          :sort-desc="[true]"
          :items-per-page="10"
          :footer-props="{'items-per-page-options': [10, 25, 50, 100]}"
          :server-items-length="getNumberSessions"
          :options.sync="pagination"
          :disable-sort="true"
        >
          <template #[`item.active`]="{ item }">
            <v-icon
              v-if="item.active"
              color="success"
            >
              check_circle
            </v-icon>
            <v-tooltip
              v-else
              bottom
            >
              <template #activator="{ on }">
                <v-icon v-on="on">
                  check_circle
                </v-icon>
              </template>
              <span>active {{ item.last_seen | lastSeen }}</span>
            </v-tooltip>
          </template>

          <template #[`item.device`]="{ item }">
            <router-link :to="{ name: 'detailsDevice', params: { id: item.device.uid } }">
              {{ item.device.name }}
            </router-link>
          </template>

          <template #[`item.username`]="{ item }">
            <v-tooltip
              v-if="!item.authenticated"
              bottom
            >
              <template #activator="{ on, attrs }">
                <span
                  v-bind="attrs"
                  v-on="on"
                >{{ item.username }}</span>
              </template>
              <span v-if="!item.authenticated">Unauthorized</span>
            </v-tooltip>
            <template
              v-if="item.authenticated"
            >
              {{ item.username }}
            </template>
          </template>

          <template #[`item.authenticated`]="{ item }">
            <v-tooltip bottom>
              <template #activator="{ on }">
                <v-icon
                  v-if="item.authenticated"
                  :color="item.active ? 'success' : ''"
                  size=""
                  v-on="on"
                >
                  mdi-shield-check
                </v-icon>
                <v-icon
                  v-else
                  color="error"
                  size=""
                  v-on="on"
                >
                  mdi-shield-alert
                </v-icon>
              </template>
              <span v-if="item.authenticated">User has been authenticated</span>
              <span v-else>User has not been authenticated</span>
            </v-tooltip>
          </template>

          <template #[`item.ip_address`]="{ item }">
            <code>{{ item.ip_address }}</code>
          </template>

          <template #[`item.started`]="{ item }">
            {{ item.started_at | formatDate }}
          </template>

          <template #[`item.last_seen`]="{ item }">
            {{ item.last_seen | formatDate }}
          </template>

          <template #[`item.actions`]="{ item }">
            <v-tooltip bottom>
              <template #activator="{ on }">
                <v-icon
                  class="icons"
                  v-on="on"
                  @click="detailsSession(item)"
                >
                  info
                </v-icon>
              </template>
              <span>Details</span>
            </v-tooltip>
            <SessionClose
              v-if="item.active && isOwner"
              :uid="item.uid"
              :device="item.device_uid"
              @update="refresh"
            />
            <SessionPlay
              :recorded="item.authenticated && item.recorded && isOwner"
              :uid="item.uid"
            />
          </template>
        </v-data-table>
      </v-card-text>
    </v-card>
  </fragment>
</template>

<script>

import SessionClose from '@/components/session/SessionClose';
import SessionPlay from '@/components/session/SessionPlay';
import { formatDate, lastSeen } from '@/components/filter/date';

export default {
  name: 'SessionList',

  components: {
    SessionClose,
    SessionPlay,
  },

  filters: { formatDate, lastSeen },

  data() {
    return {
      pagination: {},

      headers: [
        {
          text: 'Active',
          value: 'active',
          align: 'center',
        },
        {
          text: 'Device',
          value: 'device',
          align: 'center',
        },
        {
          text: 'Username',
          value: 'username',
          align: 'center',
        },
        {
          text: 'Authenticated',
          value: 'authenticated',
          align: 'center',
        },
        {
          text: 'IP Address',
          value: 'ip_address',
          align: 'center',
        },
        {
          text: 'Started',
          value: 'started',
          align: 'center',
        },
        {
          text: 'Last Seen',
          value: 'last_seen',
          align: 'center',
        },
        {
          text: 'Actions',
          value: 'actions',
          align: 'center',
        },
      ],
    };
  },

  computed: {
    getListSessions() {
      return this.$store.getters['sessions/list'];
    },

    getNumberSessions() {
      return this.$store.getters['sessions/getNumberSessions'];
    },

    isOwner() {
      return this.$store.getters['namespaces/owner'];
    },
  },

  watch: {
    pagination: {
      handler() {
        this.getSessions();
      },
      deep: true,
    },
  },

  methods: {
    detailsSession(session) {
      this.$router.push(`/session/${session.uid}`);
    },

    refresh() {
      this.getSessions();
    },

    async getSessions() {
      const data = { perPage: this.pagination.itemsPerPage, page: this.pagination.page };

      try {
        await this.$store.dispatch('sessions/fetch', data);
      } catch (e) {
        if (e.response.status === 403) {
          this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.sessionList);
        }
      }
    },
  },
};

</script>
