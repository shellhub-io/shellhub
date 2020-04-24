<template>
<fragment>

  <div class="d-flex pa-0 align-center">
    <h1 v-if="hide">Session Details</h1>
    <v-spacer/>
    <v-spacer/>
  </div>

  <v-card v-if="hide" class="mt-2">
    <v-toolbar flat color="transparent">

      <v-toolbar-title v-if="this.session.device">
        <v-icon color="success" v-if="this.session.active">check_circle</v-icon>
        <v-tooltip bottom v-else>
          <template #activator="{ on }">
            <v-icon v-on="on">check_circle</v-icon>
          </template>
          <span>active {{ session.last_seen | moment("from", "now") }}</span>
        </v-tooltip>
        {{ this.session.device.name }}
      </v-toolbar-title>

      <v-spacer></v-spacer>

      <v-icon class="icons ml-1" v-if="this.session.active" @click="closeSession()">
        desktop_access_disabled
      </v-icon>

    </v-toolbar>

    <v-divider></v-divider>

    <v-card-text>
      <div class="mt-2">
        <div class="overline">Uid</div>
        <div>{{ this.session.uid }}</div>
      </div>

      <div class="mt-2">
        <div class="overline">User</div>
        <div>{{ this.session.username }}</div>
      </div>

      <div class="mt-2">
        <div class="overline">Authenticated</div>
        <v-tooltip bottom>
            <template #activator="{ on }" v-bind:session="this.session">
              <v-icon v-on="on" :color="session.active ? 'success' : ''" size="" v-if="session.authenticated">mdi-shield-check</v-icon>
              <v-icon v-on="on" color="error" size="" v-else>mdi-shield-alert</v-icon>
            </template>
            <span v-if="session.authenticated">User has been authenticated</span>
            <span v-else>User has not been authenticated</span>
          </v-tooltip>
      </div>

      <div class="mt-2">
        <div class="overline">Ip Address</div>
        <code>{{ this.session.ip_address }}</code>
      </div>

      <div class="mt-2">
        <div class="overline">Started</div>
        <div>{{ this.session.started_at | moment("dddd, MMMM Do YYYY, h:mm:ss a") }}</div>
      </div>

      <div class="mt-2">
        <div class="overline">Last Seen</div>
        <div>{{ this.session.last_seen | moment("dddd, MMMM Do YYYY, h:mm:ss a") }}</div>
      </div>
    </v-card-text>

    <v-snackbar v-model="closeSessionSnack" :timeout=3000>Closed session conection to the device</v-snackbar>
  </v-card>

  <div class="text-center">
    <v-dialog v-model="dialog" width="500" persistent>
      <v-card>
        <v-card-title class="headline grey lighten-2" primary-title>
          Session ID error
        </v-card-title>
        <v-card-text>
        <br>
          You tried to access a non-existing session ID .
        </v-card-text>
        <v-divider></v-divider>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" text @click="redirect">
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
  name: "SessionDetails",

  data() {
    return {
      uid: "",
      session: [],
      closeSessionSnack: false,
      dialog: false,
      hide:true
    };
  },
  async created() {
    this.uid = this.$route.params.id;
    try{
      await this.$store.dispatch("sessions/get", this.uid);
      this.session = this.$store.getters["sessions/get"];
    }
    catch(error){
      this.hide=false;
      this.dialog=true;
    }
  },
  methods: {
    async closeSession() {
      this.$store.dispatch("sessions/close");
      this.closeSessionSnack = true;
      await this.$store.dispatch("sessions/get", this.uid);
      this.session = this.$store.getters["sessions/get"];
    },
    redirect(){
      this.dialog=false;
      this.$router.push('/sessions');
    }
  }
};
</script>