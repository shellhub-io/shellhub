<template>
  <fragment>
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
          {{ item.started_at | formatDateCompact }}
        </template>

        <template #[`item.last_seen`]="{ item }">
          {{ item.last_seen | formatDateCompact }}
        </template>

        <template #[`item.actions`]="{ item }">
          <v-menu
            :ref="'menu'+getListSessions.indexOf(item)"
            offset-y
          >
            <template #activator="{ on, attrs }">
              <v-chip
                color="transparent"
                v-on="on"
              >
                <v-icon
                  small
                  class="icons"
                  v-bind="attrs"
                  v-on="on"
                >
                  mdi-dots-horizontal
                </v-icon>
              </v-chip>
            </template>

            <v-card>
              <v-list-item @click.stop="detailsSession(item)">
                <v-icon left>
                  info
                </v-icon>

                <v-list-item-title>
                  Details
                </v-list-item-title>
              </v-list-item>

              <v-tooltip
                bottom
                :disabled="hasAuthorizationPlay"
              >
                <template #activator="{ on, attrs }">
                  <div
                    v-bind="attrs"
                    v-on="on"
                  >
                    <v-list-item
                      v-if="item.authenticated && item.recorded && isEnterprise"
                      :disabled="!hasAuthorizationPlay"
                      @click.stop="showSessionPlay(getListSessions.indexOf(item))"
                    >
                      <SessionPlay
                        :recorded="item.authenticated && item.recorded"
                        :uid="item.uid"
                        :show.sync="sessionPlayShow[getListSessions.indexOf(item)]"
                        data-test="sessionPlay-component"
                      />
                    </v-list-item>
                  </div>
                </template>

                <span>
                  You don't have this kind of authorization.
                </span>
              </v-tooltip>

              <v-list-item
                v-if="item.active"
                @click="showSessionClose(getListSessions.indexOf(item))"
              >
                <SessionClose
                  :uid="item.uid"
                  :device="item.device_uid"
                  :show.sync="sessionCloseShow[getListSessions.indexOf(item)]"
                  data-test="sessionClose-component"
                  @update="refresh"
                />
              </v-list-item>
            </v-card>
          </v-menu>
        </template>
      </v-data-table>
    </v-card-text>
  </fragment>
</template>

<script>

import SessionClose from '@/components/session/SessionClose';
import SessionPlay from '@/components/session/SessionPlay';
import { formatDateCompact, lastSeen } from '@/components/filter/date';
import hasPermission from '@/components/filter/permission';

export default {
  name: 'SessionListComponent',

  components: {
    SessionClose,
    SessionPlay,
  },

  filters: { formatDateCompact, lastSeen, hasPermission },

  data() {
    return {
      menu: false,
      pagination: {},
      sessionPlayShow: [],
      sessionCloseShow: [],
      playAction: 'play',
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

    isEnterprise() {
      return this.$env.isEnterprise;
    },

    hasAuthorizationPlay() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.session[this.playAction],
        );
      }

      return false;
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
      if (!this.$store.getters['boxs/getStatus']) {
        const data = { perPage: this.pagination.itemsPerPage, page: this.pagination.page };

        try {
          await this.$store.dispatch('sessions/fetch', data);

          this.setArrays();
        } catch (error) {
          if (error.response.status === 403) {
            this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
          } else {
            this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.sessionList);
          }
        }
      } else {
        this.setArrays();
        this.$store.dispatch('boxs/setStatus', false);
      }
    },

    showSessionPlay(index) {
      this.sessionPlayShow[index] = this.sessionPlayShow[index] === undefined
        ? true : !this.sessionPlayShow[index];
      this.$set(this.sessionPlayShow, index, this.sessionPlayShow[index]);

      this.closeMenu(index);
    },

    showSessionClose(index) {
      this.sessionCloseShow[index] = this.sessionCloseShow[index] === undefined
        ? true : !this.sessionCloseShow[index];
      this.$set(this.sessionCloseShow, index, this.sessionCloseShow[index]);

      this.closeMenu(index);
    },

    setArrays() {
      const numberSessions = this.getListSessions.length;

      if (numberSessions > 0) {
        this.sessionPlayShow = new Array(this.getNumberSessions).fill(false);
        this.sessionCloseShow = new Array(this.getNumberSessions).fill(false);
      }
    },

    closeMenu(index) {
      this.$refs[`menu${index}`].isActive = false;
    },
  },
};

</script>
