<template>
  <fragment>
    <div class="d-flex pa-0 align-center">
      <h1 v-if="hide">
        Session Details
      </h1>
      <v-spacer />
      <v-spacer />
    </div>

    <v-card
      v-if="session"
      class="mt-2"
    >
      <v-toolbar
        flat
        color="transparent"
      >
        <v-toolbar-title v-if="session.device">
          <v-icon
            v-if="session.active"
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
            <span>active {{ session.last_seen | lastSeen }}</span>
          </v-tooltip>
          {{ session.device.name }}
        </v-toolbar-title>

        <v-spacer />

        <SessionPlay
          :uid="session.uid"
          :recorded="session.authenticated && session.recorded && isOwner"
        />

        <SessionClose
          v-if="session.active"
          :uid="session.uid"
          :device="session.device_uid"
          @update="refresh"
        />
      </v-toolbar>

      <v-divider />

      <v-card-text>
        <div class="mt-2">
          <div class="overline">
            Uid
          </div>
          <div
            data-test="sessionUid-field"
          >
            {{ session.uid }}
          </div>
        </div>

        <div class="mt-2">
          <div class="overline">
            User
          </div>
          <div
            data-test="sessionUser-field"
          >
            {{ session.username }}
          </div>
        </div>

        <div class="mt-2">
          <div
            class="overline"
          >
            Authenticated
          </div>
          <v-tooltip bottom>
            <template
              #activator="{ on }"
              :session="session"
            >
              <v-icon
                v-if="session.authenticated"
                :color="session.active ? 'success' : ''"
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
            <span v-if="session.authenticated">User has been authenticated</span>
            <span v-else>User has not been authenticated</span>
          </v-tooltip>
        </div>

        <div class="mt-2">
          <div class="overline">
            Ip Address
          </div>
          <code
            data-test="sessionIpAddress-field"
          >
            {{ session.ip_address }}
          </code>
        </div>

        <div class="mt-2">
          <div class="overline">
            Started
          </div>
          <div
            data-test="sessionStartedAt-field"
          >
            {{ session.started_at | formatDate }}
          </div>
        </div>

        <div class="mt-2">
          <div class="overline">
            Last Seen
          </div>
          <div
            data-test="sessionLastSeen-field"
          >
            {{ session.last_seen | formatDate }}
          </div>
        </div>
      </v-card-text>
    </v-card>

    <div class="text-center">
      <v-dialog
        v-model="dialog"
        width="500"
        persistent
      >
        <v-card>
          <v-card-title
            class="headline grey lighten-2"
            primary-title
          >
            Session ID error
          </v-card-title>
          <v-card-text class="mt-4 mb-3 pb-1">
            You tried to access a non-existing session ID
          </v-card-text>
          <v-card-actions>
            <v-spacer />
            <v-btn
              color="primary"
              text
              @click="redirect"
            >
              Go back to sessions
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
    </div>
  </fragment>
</template>

<script>

import SessionPlay from '@/components/session/SessionPlay';
import SessionClose from '@/components/session/SessionClose';
import { formatDate, lastSeen } from '@/components/filter/date';

export default {
  name: 'SessionDetails',

  components: {
    SessionPlay,
    SessionClose,
  },

  filters: { formatDate, lastSeen },

  data() {
    return {
      uid: '',
      session: null,
      dialog: false,
      hide: true,
    };
  },

  computed: {
    isOwner() {
      return this.$store.getters['namespaces/owner'];
    },
  },

  async created() {
    this.uid = this.$route.params.id;
    try {
      await this.$store.dispatch('sessions/get', this.uid);
      this.session = this.$store.getters['sessions/get'];
    } catch (error) {
      this.hide = false;
      this.dialog = true;
      this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.sessionDetails);
    }
  },

  methods: {
    redirect() {
      this.dialog = false;
      this.$router.push('/sessions');
    },

    async refresh() {
      try {
        await this.$store.dispatch('sessions/get', this.uid);
        this.session = this.$store.getters['sessions/get'];
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.sessionDetails);
      }
    },
  },
};

</script>
