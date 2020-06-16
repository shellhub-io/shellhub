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
            <span>active {{ lastActive }}</span>
          </v-tooltip>
          {{ device.name }}
        </v-toolbar-title>
        <DeviceRename
          :device="device"
          @newHostname="receiveName"
        />

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
          <div>{{ convertDate }}</div>
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
          <v-card-text class="mt-4 mb-3 pb-1">
            You tried to access a non-existing device ID
          </v-card-text>
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

import moment from 'moment';
import TerminalDialog from '@/components/terminal/TerminalDialog';
import DeviceIcon from '@/components/device/DeviceIcon';
import DeviceDelete from '@/components/device/DeviceDelete';
import DeviceRename from '@/components/device/DeviceRename';

export default {
  name: 'DeviceDetails',

  components: {
    TerminalDialog,
    DeviceIcon,
    DeviceDelete,
    DeviceRename,
  },

  data() {
    return {
      dialogDelete: false,
      dialogError: false,
      uid: '',
      hostname: window.location.hostname,
      editName: '',
      hide: true,
      device: null,
    };
  },

  computed: {
    lastActive() {
      return moment(this.device.last_seen).format('from', 'now');
    },

    convertDate() {
      return moment(this.device.last_seen).format('dddd, MMMM Do YYYY, h:mm:ss a');
    },
  },

  async created() {
    this.uid = await this.$route.params.id;
    try {
      await this.$store.dispatch('devices/get', this.uid);
      this.device = this.$store.getters['devices/get'];
    } catch (error) {
      this.hide = false;
      this.dialogError = true;
    }
  },

  methods: {
    redirect() {
      this.dialogError = false;
      this.$router.push('/devices');
    },

    receiveName(params) {
      this.device.name = params;
    },
  },
};

</script>
