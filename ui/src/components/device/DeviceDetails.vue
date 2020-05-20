<template>
  <fragment>
    <div class="d-flex pa-0 align-center">
      <h1 v-if="hide">
        Device Details
      </h1>
    </div>
  
    <v-card
      v-if="device"
      class="mt-2"
    >
      <v-toolbar
        flat
        color="transparent"
      >
        <v-edit-dialog
          :return-value="editName"
          large
          @open="editName = device.name"
          @save="save()"
        >
          <v-text-field
            slot="input"
            v-model="editName"
            label="Edit"
            single-line
          />
          <v-toolbar-title>
            <v-icon
              v-if="device.online"
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
              <span>active {{ device.last_seen | moment("from", "now") }}</span>
            </v-tooltip>
            {{ device.name }}
            <v-icon
              small
              left
            >
              mdi-file-edit
            </v-icon>
          </v-toolbar-title>
        </v-edit-dialog>
      
        <v-spacer />
      
        <TerminalDialog :uid="device.uid" />
      
        <DeviceDelete
          :uid="device.uid"
          :dialog="dialogDelete"
          :redirect="true"
        />
      </v-toolbar>

      <v-divider />

      <v-card-text>
        <div class="mt-2">
          <div class="overline">
            UID
          </div>
          <div>{{ device.uid }}</div>
        </div>
        
        <div class="mt-2">
          <div class="overline">
            MAC
          </div>
          <code v-if="device.identity">{{ device.identity['mac'] }}</code>
        </div>

        <div class="mt-2">
          <div class="overline">
            Operating System
          </div>
          <div v-if="device.info">
            <DeviceIcon :icon-name="device.info.id" />
            {{ device.info.pretty_name }}
          </div>
        </div>

        <div class="mt-2">
          <div class="overline">
            Last Seen
          </div>
          <div>{{ device.last_seen | moment("dddd, MMMM Do YYYY, h:mm:ss a") }}</div>
        </div>
      </v-card-text>
    </v-card>

    <div class="text-center">
      <v-dialog
        v-model="dialogError"
        persistent
        width="500"
      >
        <v-card>
          <v-card-title
            class="headline grey lighten-2"
            primary-title
          >
            Device ID error
          </v-card-title>
          <v-card-text>
            <br>
            You tried to access a non-existing device ID.
          </v-card-text>
          <v-divider />
          <v-card-actions>
            <v-spacer />
            <v-btn
              color="primary"
              text
              @click="redirect"
            >
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
import DeviceIcon from '@/components/device/DeviceIcon.vue';
import DeviceDelete from '@/components/device/DeviceDelete.vue'; 

export default {
  name: 'DeviceDetails',

  components: {
    TerminalDialog,
    DeviceIcon,
    DeviceDelete,
  },

  data() {
    return {
      dialogDelete: false,
      dialogError:false,
      uid: '',
      hostname: window.location.hostname,
      editName: '',
      hide:true,
      device: null,
    };
  },

  async created() {
    this.uid = await this.$route.params.id;
    try{
      await this.$store.dispatch('devices/get', this.uid);
      this.device = this.$store.getters['devices/get'];
    } catch(error){
      this.hide=false;
      this.dialogError=true;
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
    formatDate() {
      return moment(String(this.device.last_seen)).format('DD-MM-YYYY');
    },
    redirect(){
      this.dialogError=false;
      this.$router.push('/devices');
    }
  },
};
</script>
