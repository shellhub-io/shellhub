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
            <span>active {{ session.last_seen | moment("from", "now") }}</span>
          </v-tooltip>
          {{ session.device.name }}
        </v-toolbar-title>

        <v-spacer />

        <v-icon
          v-if="session.active"
          class="icons ml-1"
          @click="closeSession()"
        >
          desktop_access_disabled
        </v-icon>
      </v-toolbar>

      <v-divider />

      <v-card-text>
        <div class="mt-2">
          <div class="overline">
            Uid
          </div>
          <div>{{ session.uid }}</div>
        </div>

        <div class="mt-2">
          <div class="overline">
            User
          </div>
          <div>{{ session.username }}</div>
        </div>

        <div class="mt-2">
          <div class="overline">
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
          <code>{{ session.ip_address }}</code>
        </div>

        <div class="mt-2">
          <div class="overline">
            Started
          </div>
          <div>{{ session.started_at | moment("dddd, MMMM Do YYYY, h:mm:ss a") }}</div>
        </div>

        <div class="mt-2">
          <div class="overline">
            Last Seen
          </div>
          <div>{{ session.last_seen | moment("dddd, MMMM Do YYYY, h:mm:ss a") }}</div>
        </div>
      </v-card-text>

      <v-snackbar
        v-model="closeSessionSnack"
        :timeout="3000"
      >
        Closed session conection to the device
      </v-snackbar>
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
          <v-card-text>
            <br>
            You tried to access a non-existing session ID .
          </v-card-text>
          <v-divider />
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
export default {
  name: 'SessionDetails',

  data() {
    return {
      uid: '',
      session: null,
      closeSessionSnack: false,
      dialog: false,
      hide:true
    };
  },
  async created() {
    this.uid = this.$route.params.id;
    try{
      await this.$store.dispatch('sessions/get', this.uid);
      this.session = this.$store.getters['sessions/get'];
    } catch(error){
      this.hide=false;
      this.dialog=true;
    }
  },
  methods: {
    async closeSession() {
      this.$store.dispatch('sessions/close');
      this.closeSessionSnack = true;
      await this.$store.dispatch('sessions/get', this.uid);
      this.session = this.$store.getters['sessions/get'];
    },
    redirect(){
      this.dialog=false;
      this.$router.push('/sessions');
    }
  }
};
</script>