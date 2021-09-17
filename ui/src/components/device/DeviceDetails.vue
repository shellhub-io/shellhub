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
            <span>active {{ device.last_seen | lastSeen }}</span>
          </v-tooltip>
          {{ device.name }}
        </v-toolbar-title>
        <DeviceRename
          :name="device.name"
          :uid="device.uid"
          data-test="deviceRename-component"
          @newHostname="receiveName"
        />

        <v-spacer />

        <TerminalDialog
          v-if="device.online"
          :uid="device.uid"
          data-test="terminalDialog-component"
        />

        <DeviceDelete
          :uid="device.uid"
          :dialog="dialogDelete"
          :redirect="true"
          data-test="deviceDelete-component"
        />
      </v-toolbar>

      <v-divider />

      <v-card-text>
        <div class="mt-2">
          <div class="overline">
            UID
          </div>
          <div
            data-test="deviceUid-field"
          >
            {{ device.uid }}
          </div>
        </div>

        <div class="mt-2">
          <div class="overline">
            MAC
          </div>
          <code
            v-if="device.identity"
            data-test="deviceMac-field"
          >
            {{ device.identity['mac'] }}
          </code>
        </div>

        <div class="mt-2">
          <v-combobox
            id="targetInput"
            v-model="list"
            label="Tag"
            hint="Maximum of 5 tags"
            multiple
            chips
            append-icon
            data-test="deviceTag-field"
            :deletable-chips="true"
            :rules="[tagRule]"
          />
        </div>

        <div class="mt-2">
          <div class="overline">
            Operating System
          </div>
          <div
            v-if="device.info"
            data-test="devicePrettyName-field"
          >
            <DeviceIcon :icon-name="device.info.id" />
            {{ device.info.pretty_name }}
          </div>
        </div>

        <div class="mt-2">
          <div class="overline">
            Last Seen
          </div>
          <div
            data-test="deviceConvertDate-field"
          >
            {{ device.last_seen | formatDate }}
          </div>
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
            You tried to access a non-existing device ID.
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

import TerminalDialog from '@/components/terminal/TerminalDialog';
import DeviceIcon from '@/components/device/DeviceIcon';
import DeviceDelete from '@/components/device/DeviceDelete';
import DeviceRename from '@/components/device/DeviceRename';
import { formatDate, lastSeen } from '@/components/filter/date';

export default {
  name: 'DeviceDetails',

  components: {
    TerminalDialog,
    DeviceIcon,
    DeviceDelete,
    DeviceRename,
  },

  filters: { formatDate, lastSeen },

  data() {
    return {
      uid: '',
      errorMsg: '',
      hostname: window.location.hostname,
      hide: true,
      device: null,
      dialogDelete: false,
      dialogError: false,
      list: [],
      oldList: [],
    };
  },

  watch: {
    list(newList) {
      if (JSON.stringify(newList) !== JSON.stringify(this.oldList)) {
        this.actionTag(newList);
      }
    },
  },

  async created() {
    this.uid = await this.$route.params.id;
    try {
      await this.$store.dispatch('devices/get', this.uid);
      this.device = this.$store.getters['devices/get'];
      this.list = this.device.tags;
      this.oldList = this.device.tags;
    } catch (error) {
      this.hide = false;
      this.dialogError = true;
      this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.deviceDetails);
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

    tagRule() {
      if (this.errorMsg !== '') {
        return this.errorMsg;
      }

      return true;
    },

    async actionTag(newList) {
      const device = this.$store.getters['devices/get'];
      const data = { uid: device.uid, tags: newList };

      try {
        this.errorMsg = '';
        await this.$store.dispatch('devices/updateDeviceTag', data);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.deviceTagUpdate);
        this.oldList = newList;
      } catch (error) {
        this.$nextTick(() => this.list.pop());
        switch (true) {
        // when the name the format is invalid.
        case (error.response.status === 400): {
          this.errorMsg = 'The format is invalid. Min 3, Max 255 characters!';
          break;
        }
        // when the user is not authorized.
        case (error.response.status === 403): {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.deviceTagUpdate);
          break;
        }
        // When the array tag size reached the max capacity.
        case (error.response.status === 406): {
          this.errorMsg = 'The maximum capacity has reached.';
          break;
        }
        default: {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.deviceTagUpdate);
        }
        }
      }
      return false;
    },
  },
};

</script>
