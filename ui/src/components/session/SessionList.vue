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
          :items="listSessions"
          item-key="uid"
          :sort-by="['started_at']"
          :sort-desc="[true]"
          :items-per-page="10"
          :footer-props="{'items-per-page-options': [10, 25, 50, 100]}"
          :server-items-length="numberSessions"
          :options.sync="pagination"
          :disable-sort="true"
        >
          <template v-slot:item.active="{ item }">
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
              <span>active {{ item.last_seen | moment("from", "now") }}</span>
            </v-tooltip>
          </template>

          <template v-slot:item.device="{ item }">
            <router-link :to="{ name: 'detailsDevice', params: { id: item.device.uid } }">
              {{ item.device.name }}
            </router-link>
          </template>

          <template v-slot:item.username="{ item }">
            <v-tooltip
              v-if="!item.authenticated"
              bottom
            >
              <template v-slot:activator="{ on, attrs }">
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

          <template v-slot:item.authenticated="{ item }">
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

          <template v-slot:item.ip_address="{ item }">
            <code>{{ item.ip_address }}</code>
          </template>

          <template v-slot:item.started="{ item }">
            {{ item.started_at | moment("ddd, MMM Do YY, h:mm:ss a") }}
          </template>

          <template v-slot:item.last_seen="{ item }">
            {{ item.last_seen | moment("ddd, MMM Do YY, h:mm:ss a") }}
          </template>

          <template v-slot:item.actions="{ item }">
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
              v-if="item.active"
              :session="item"
              @update="refresh"
            />
          </template>
        </v-data-table>
      </v-card-text>
      <v-snackbar
        v-model="sessionSnack"
        :timeout="3000"
      >
        Session closed
      </v-snackbar>
    </v-card>
  </fragment>
</template>

<script>

import SessionClose from '@/components/session/SessionClose';

export default {
  name: 'SessionList',

  components: {
    SessionClose,
  },

  data() {
    return {
      sessionSnack: false,
      numberSessions: 0,
      listSessions: [],
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
      await this.$store.dispatch('sessions/fetch', data);
      this.listSessions = this.$store.getters['sessions/list'];
      this.numberSessions = this.$store.getters['sessions/getNumberSessions'];
    },
  },
};

</script>
