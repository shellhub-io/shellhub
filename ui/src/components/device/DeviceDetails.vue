<template>
<fragment>

  <div class="d-flex pa-0 align-center">
    <h1 v-if="hide">Device Details</h1>
  </div>
  
  <v-card class="mt-2" v-if="hide">
    <v-toolbar flat color="transparent">
      
      <v-edit-dialog :return-value="editName" large @open="editName = device.name" @save="save()">
        <v-text-field slot="input" v-model="editName" label="Edit" single-line></v-text-field>
        <v-toolbar-title>
          <v-icon color="success" v-if="this.device.online">check_circle</v-icon>
          <v-tooltip bottom v-else>
            <template #activator="{ on }">
              <v-icon v-on="on">check_circle</v-icon>
            </template>
            <span>active {{ device.last_seen | moment("from", "now") }}</span>
          </v-tooltip>
          {{ this.device.name }}
          <v-icon small left>mdi-file-edit</v-icon>
        </v-toolbar-title>
      </v-edit-dialog>
      
      <v-spacer></v-spacer>
      
      <TerminalDialog :uid="device.uid"></TerminalDialog>
      
      <v-btn icon>
        <v-icon @click="remove()">delete</v-icon>
      </v-btn>
    </v-toolbar>

      <v-divider></v-divider>

      <v-card-text>
        <div class="mt-2">
          <div class="overline">UID</div>
          <div>{{ this.device.uid }}</div>
        </div>
        
        <div class="mt-2">
          <div class="overline">MAC</div>
          <code v-if="this.device.identity">{{ this.device.identity['mac'] }}</code>
        </div>

        <div class="mt-2">
          <div class="overline">Operating System</div>
          <div v-if="this.device.info">
            <v-icon left>{{ deviceIcon[device.info.id] || 'fl-tux' }}</v-icon>
            {{ this.device.info.pretty_name }}
          </div>
        </div>

        <div class="mt-2">
          <div class="overline">Last Seen</div>
          <div>{{ this.device.last_seen | moment("dddd, MMMM Do YYYY, h:mm:ss a") }}</div>
        </div>
      </v-card-text>

    </v-card>

  <div class="text-center">
    <v-dialog persistent v-model="dialog" width="500">
      <v-card>
        <v-card-title class="headline grey lighten-2" primary-title>
          Device ID error
        </v-card-title>
        <v-card-text>
        <br>
          You tried to access a non-existing device ID.
        </v-card-text>
        <v-divider></v-divider>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" text @click="redirect">
            Go back to devices
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</fragment>
</template>

<script>

import TerminalDialog from '@/components/terminal/TerminalDialog.vue';
import moment from 'moment';

export default {
  name: 'DeviceDetails',

  components: {
    TerminalDialog
  },

  data() {
    return {
      dialog:false,
      uid: '',
      hostname: window.location.hostname,
      editName: '',
      hide:true,
      device: [],
      deviceIcon: {
        arch: 'fl-archlinux',
        ubuntu: 'fl-ubuntu'
      },
    };
  },

  async created() {
    this.uid = this.$route.params.id;
    try{
      await this.$store.dispatch('devices/get', this.uid);
      this.device = this.$store.getters["devices/get"];
    }
    catch(error){
      this.hide=false;
      this.dialog=true;
    } 
  },
  methods: {
    save() {
      this.$store.dispatch('devices/rename', {
        uid: this.device.uid,
        name: this.editName
      });

      this.device.name = this.editName;
    },
    remove() {
      if (confirm('Are you sure?')) {
        this.$store.dispatch('devices/remove', this.device.uid);
      }
    },
    format_date() {
      return moment(String(this.device.last_seen)).format('DD-MM-YYYY');
    },
    redirect(){
      this.dialog=false;
      this.$router.push('/devices');
    }
  },

};
</script>
