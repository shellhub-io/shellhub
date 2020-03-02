<template>
<fragment>

  <div class="d-flex pa-0 align-center">
  <h1>Session Details</h1>
  <v-spacer/>
  <v-spacer/>
  <!-- <AddDevice/> -->
  <!-- <v-btn outlined @click="$store.dispatch('modals/showAddDevice', true)">Add Device</v-btn> -->
  </div>
    <v-card class="mt-2">
        <v-app-bar flat color="transparent">

          <div class="item-title">
              
              <div class="item-title-name">
                  {{ this.session.device }}
              </div>

              <div class="status">
                <div class="status-online" v-if="this.session.active">
                  Online
                </div>
                <div class="status-offline" v-else>
                  Offline
                </div>
              </div>
                    
          </div>

        </v-app-bar>
        

        <v-divider></v-divider>
        <v-card-text>
          
          <div class="item">
            <div class="item-name">Uid</div>
            <div class="item-description">{{this.session.uid}}</div>
          </div>

          <div class="item">
            <div class="item-name">User</div>
            <div class="item-description">{{this.session.username}}</div>
          </div>

          <div class="item"> 
            <div class="item-name">Ip Address</div>
            <div class="item-description">{{this.session.ip_address}}</div>
          </div>

          <div class="item">
            <div class="item-name">Started</div>
            <div class="item-description">{{this.session.started_at | moment("dddd, MMMM Do YYYY, h:mm:ss a")}}</div>
          </div>

          <div class="item">
            <div class="item-name">Last Seen</div>
            <div class="item-description">{{this.session.last_seen | moment("dddd, MMMM Do YYYY, h:mm:ss a")}}</div>
          </div>

        </v-card-text>
        <v-snackbar v-model="copySnack" :timeout=3000>Device SSHID copied to clipboard</v-snackbar>
    </v-card>
</fragment>
</template>

<script>
// import moment from 'moment'

export default {
  name: "DetailsSession",

  components: {
  },

  async created() {
      this.uid = this.$route.params.id
      await this.$store.dispatch("sessions/get", this.uid);
      this.session = this.$store.getters["sessions/get"];
  },

  computed: {
  },

  methods: {
  },

  data() {
    return {
      uid: '',
      session: [],
      copySnack: false
    };
  }
};
</script>
<style scoped>

.mt-2{
  /* position: relative; */
  width:100%;
}

.item {
  margin-left: 15px;
  margin-right: 15px;
  margin-bottom: 15px;
  width: 100%;
}
.item-name {
  font-size: 14px;
  color: #FFFFFF; 
  display:inline;
}

.item-description {
  font-size: 14px;
  /* display:inline; */
}

.item-title{
  margin-bottom: -4px;
  margin-left: 15px;
  width: 350%;
}

.item-title-name{
  font-size: 28px;
  font-weight: 500;
  display:inline;
}

.status{
  margin-top: -5px;
  font-size: 12px;
  font-weight: bold;
}
.status-online{
  color: rgb(162, 250, 163);
}

.status-offline{
  color: red;
}

.item-action{
  margin-left: 10px;
  /* display:inline; */
  width: 20%;
  float: right;
}

</style>