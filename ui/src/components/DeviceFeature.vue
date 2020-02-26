<template>
<fragment>

  <div class="d-flex pa-0 align-center">
  <h1>Device Feature</h1>
  <v-spacer/>
  <v-spacer/>
  <!-- <AddDevice/> -->
  <!-- <v-btn outlined @click="$store.dispatch('modals/showAddDevice', true)">Add Device</v-btn> -->
  </div>
    <v-card class="mt-2">
        <v-app-bar flat color="transparent">
        </v-app-bar>
        <v-divider></v-divider>
        <v-card-text>
          
          <div class="item">
            <div class="item-name">Uid: </div>
            <div class="item-description">{{this.device.uid}}</div>
          </div>

          <div class="item">
            <div class="item-description">

              <v-edit-dialog :return-value="editName" large @open="editName = device.name" @save="save()">
                  <div class="item-name">Name: </div>
                  
                  <v-text-field slot="input" v-model="editName" label="Edit" single-line>
                  </v-text-field>
                  
                  <v-icon small left>mdi-file-edit</v-icon>
                  <div class="item-description">
                      {{ this.device.name }}
                  </div>

              </v-edit-dialog>

            </div> 
          </div>

          <div class="item">
            <div class="item-name">Mac: </div>
            <div class="item-description">{{this.device.identity}}</div>
          </div>

          <div class="item"> 
            <div class="item-name">Operating System: </div>
            <div class="item-description">{{this.device.attributes}}</div>
          </div>

          <div class="item">
            <div class="item-name">Public Key: </div>
            <div>{{this.device.public_key}}</div>
          </div>

          <div class="item">
            <div class="item-name">Tenant Id: </div>
            <div class="item-description">{{this.device.tenant_id}}</div>
          </div>

          <div class="item">
            <div class="item-name">Last Seen: </div>
            <div class="item-description">{{this.device.last_seen}}</div>
          </div>

          <div class="item">
            <div class="item-name">Online: </div>
            <div class="item-description">{{this.device.online}}</div>
          </div>

          <div class="item">
            <div class="item-name">Namespace: </div>
            <div class="item-description">{{this.device.namespace}}</div>
          </div>

        </v-card-text>
        <v-snackbar v-model="copySnack" :timeout=3000>Device SSHID copied to clipboard</v-snackbar>
    </v-card>
</fragment>
</template>

<script>

export default {
  name: "DeviceFeature",

  components: {

  },

  async created() {
      this.uid = this.$route.params.id
      await this.$store.dispatch("devices/get", this.uid);
      this.device = this.$store.getters["devices/get"];

  },

  computed: {
    // device() {
    //   return this.$store.getters["devices/get"];
    // }
  },

  methods: {

    save() {
      this.$store.dispatch("devices/rename", {
        uid: this.device.uid,
        name: this.editName
      });

      this.device.name = this.editName;
    }
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

.item {
  margin-bottom: 4px;
}
.item-name {
  font-size: 14px;
  color: #FFFFFF; 
  display:inline;
}

.item-description {
  font-size: 14px;
  display:inline;

}



</style>