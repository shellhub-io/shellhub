<template>
<fragment>

  <div class="d-flex pa-0 align-center">
  <h1>Device Details</h1>
  <v-spacer/>
  <v-spacer/>
  <!-- <AddDevice/> -->
  <!-- <v-btn outlined @click="$store.dispatch('modals/showAddDevice', true)">Add Device</v-btn> -->
  </div>
    <v-card class="mt-2">
        <v-app-bar flat color="transparent">
          <div class="item-title">
          
            <!-- <div class="item-description"> -->
              <v-edit-dialog :return-value="editName" large @open="editName = device.name" @save="save()">
                  
                  <v-text-field slot="input" v-model="editName" label="Edit" single-line>
                  </v-text-field>
                  
                  <div class="item-title-name">
                      {{ this.device.name }}
                  </div>

                  <v-icon small left>mdi-file-edit</v-icon>

                  <div class="status">
                    <div class="status-online" v-if="this.device.online">
                      Online
                    </div>
                    <div class="status-offline" v-else>
                      Offline
                    </div>
                  </div>

              </v-edit-dialog>
            <!-- </div>  -->
        
          </div>

          <div class="item-action">

            <TerminalDialog :uid="device.uid"></TerminalDialog>

            <v-icon @click="remove()">
                delete
            </v-icon>
          </div>

        </v-app-bar>
        

        <v-divider></v-divider>
        <v-card-text>
          
          <div class="item">
            <div class="item-name">Uid </div>
            <div class="item-description">{{this.device.uid}}</div>
          </div>

          <div class="item">
            <div class="item-name">Mac </div>
            <div class="item-description" v-if="this.device.identity">{{this.device.identity['mac']}}</div>
          </div>

          <div class="item"> 
            <div class="item-name">Operating System </div>
            <div class="item-description" v-if="this.device.attributes">{{this.device.attributes.pretty_name}}</div>
          </div>

          <div class="item">
            <div class="item-name">Last Seen </div>
            <div class="item-description">{{this.device.last_seen | moment("dddd, MMMM Do YYYY, h:mm:ss a")}}</div>
          </div>

        </v-card-text>
        <v-snackbar v-model="copySnack" :timeout=3000>Device SSHID copied to clipboard</v-snackbar>
    </v-card>
</fragment>
</template>

<script>
import TerminalDialog from "@/components/TerminalDialog.vue";
import moment from 'moment'

export default {
  name: "DetailsDevice",

  components: {
    TerminalDialog,
  },

  async created() {
      this.uid = this.$route.params.id
      await this.$store.dispatch("devices/get", this.uid);
      this.device = this.$store.getters["devices/get"];
  },

  computed: {
  },

  methods: {

    save() {
      this.$store.dispatch("devices/rename", {
        uid: this.device.uid,
        name: this.editName
      });

      this.device.name = this.editName;
    },

    remove() {
      if (confirm("Are you sure?")) {
        this.$store.dispatch("devices/remove", this.device.uid);
        this.$router.push('/devices');
      }
    },
    format_date(){
        return moment(String(this.device.last_seen)).format('DD-MM-YYYY')
    },
  },

  data() {
    return {
      uid: '',
      hostname: window.location.hostname,
      copySnack: false,
      editName: "",
      device: []
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