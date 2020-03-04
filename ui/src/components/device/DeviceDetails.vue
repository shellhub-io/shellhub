<template>
<fragment>

  <div class="d-flex pa-0 align-center">
    <h1>Device Details</h1>
  </div>
  
  <v-card class="mt-2">
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
          <div v-if="this.device.attributes">
            <v-icon left>{{ deviceIcon[device.attributes.id] || 'fl-tux' }}</v-icon>
            {{ this.device.attributes.pretty_name }}
          </div>
        </div>

        <div class="mt-2">
          <div class="overline">Last Seen</div>
          <div>{{ this.device.last_seen | moment("dddd, MMMM Do YYYY, h:mm:ss a") }}</div>
        </div>
      </v-card-text>

    </v-card>
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
      uid: '',
      hostname: window.location.hostname,
      editName: '',
      device: [],
      deviceIcon: {
        arch: 'fl-archlinux',
        ubuntu: 'fl-ubuntu'
      },
    };
  },

  async created() {
    this.uid = this.$route.params.id;
    await this.$store.dispatch('devices/get', this.uid);
    this.device = this.$store.getters['devices/get'];
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
        this.$router.push('/devices');
      }
    },
    format_date() {
      return moment(String(this.device.last_seen)).format('DD-MM-YYYY');
    }
  },

};
</script>
