<template>
<fragment>

  <div class="d-flex pa-0 align-center">
    <h1>Sessions</h1>
    <v-spacer/>
    <v-spacer/>
  </div>

  <v-card class="mt-2">
    <v-app-bar flat color="transparent">
      <v-toolbar-title></v-toolbar-title>
    </v-app-bar>

    <v-divider></v-divider>

    <!-- v-icon notranslate v-icon--link mdi mdi-content-copy theme--light -->

    <v-card-text class="pa-0">
      <v-data-table :headers="headers" :items=listSessions item-key="uid" :sort-by="['started_at']" :sort-desc="[true]" disable-pagination hide-default-footer>

        <template v-slot:item.active="{ item }">
          <v-icon color="success" v-if="item.active">check_circle</v-icon>
          <v-tooltip bottom v-else>
            <template #activator="{ on }">
              <v-icon v-on="on">check_circle</v-icon>
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
          <v-tooltip bottom v-if="!item.authenticated">
            <template #activator="{ on }">
              <v-icon v-on="on" color="error" size="medium">mdi-shield-alert</v-icon>
            </template>
            <span>Unauthorized</span>
          </v-tooltip>
          {{ item.username }}
        </template>

        <template v-slot:item.ip_address="{ item }">
          <code>{{ item.ip_address }}</code>
        </template>

        <template v-slot:item.started="{ item }">
          {{ item.started_at | moment("ddd, MMM Do YY, h:mm:ss a")}}
        </template>

        <template v-slot:item.last_seen="{ item }">
          {{ item.last_seen | moment("ddd, MMM Do YY, h:mm:ss a")}}
        </template>

        <template v-slot:item.actions="{ item }">
          <v-icon class="icons" @click="detailsSession(item)">
            info
          </v-icon>

          <v-icon class="icons ml-1" v-if="item.active" @click="closeSession(item)">
            desktop_access_disabled
          </v-icon>
        </template>
      </v-data-table>
    </v-card-text>
  <v-snackbar v-model="sessionSnack" :timeout=3000>Closed session conection to the device</v-snackbar>
  </v-card>

</fragment>
</template>

<script>
export default {
  name: "SessionList",

  data() {
    return {
      sessionSnack: false,
      listSessions: [],

      headers: [
        {
          text: "Active",
          value: "active",
          align: "center"
        },
        {
          text: "Device",
          value: "device",
          align: "center"
        },
        {
          text: "Username",
          value: "username",
          align: "center"
        },
        {
          text: "IP Address",
          value: "ip_address",
          align: "center"
        },
        {
          text: "Started",
          value: "started",
          align: "center"
        },
        {
          text: "Last Seen",
          value: "last_seen",
          align: "center"
        },
        {
          text: "Actions",
          value: "actions",
          align: "center"
        }
      ]
    };
  },

  async mounted() {
    await this.$store.dispatch("sessions/fetch");
    this.listSessions = this.$store.getters["sessions/list"];
  },

  methods: {
    detailsSession(session) {
      this.$router.push("/session/" + session.uid);
    },
    async closeSession(session) {
      this.$store.dispatch("sessions/close", session);
      this.sessionSnack = true;

      await this.$store.dispatch("sessions/fetch");
      this.listSessions = this.$store.getters["sessions/list"];
    }
  }
};
</script>
