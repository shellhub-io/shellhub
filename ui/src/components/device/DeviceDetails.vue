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
          <TerminalDialog
            v-if="device.status === 'accepted'"
            :enable-connect-button="true"
            :uid="device.uid"
            :online="device.online"
            data-test="terminalDialog-component"
          />

          {{ device.name }}
        </v-toolbar-title>

        <v-spacer />

        <v-menu
          ref="menu"
          offset-y
        >
          <template #activator="{ on, attrs }">
            <v-chip
              color="transparent"
              v-on="on"
            >
              <v-icon
                small
                class="icons"
                v-bind="attrs"
                v-on="on"
              >
                mdi-dots-horizontal
              </v-icon>
            </v-chip>
          </template>

          <v-card>
            <v-list-item @click.stop="openDialog('deviceRenameShow')">
              <DeviceRename
                :name="device.name"
                :uid="device.uid"
                :show.sync="deviceRenameShow"
                data-test="deviceRename-component"
                @newHostname="receiveName"
              />
            </v-list-item>

            <v-tooltip
              bottom
              :disabled="hasAuthorizationFormUpdate"
            >
              <template #activator="{ on, attrs }">
                <div
                  v-bind="attrs"
                  v-on="on"
                >
                  <v-list-item
                    :disabled="!hasAuthorizationFormUpdate"
                    @click.stop="openDialog('tagFormUpdateShow')"
                  >
                    <TagFormUpdate
                      :device-uid="device.uid"
                      :tags-list="device.tags"
                      :show.sync="tagFormUpdateShow"
                      data-test="tagFormUpdate-component"
                      @update="getDevice()"
                    />
                  </v-list-item>
                </div>
              </template>

              <span>
                You don't have this kind of authorization.
              </span>
            </v-tooltip>

            <v-list-item @click.stop="openDialog('deviceDeleteShow')">
              <DeviceDelete
                :uid="device.uid"
                :redirect="true"
                :show.sync="deviceDeleteShow"
                data-test="deviceDelete-component"
              />
            </v-list-item>
          </v-card>
        </v-menu>
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
            Tags
          </div>

          <v-chip
            v-for="(tag, index) in device.tags"
            :key="index"
            class="ml-1 mb-1"
            small
            outlined
          >
            {{ tag }}
          </v-chip>
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
          <v-card-title class="headline primary">
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
import TagFormUpdate from '@/components/tag/TagFormUpdate';
import { formatDate, lastSeen } from '@/components/filter/date';
import hasPermission from '@/components/filter/permission';

export default {
  name: 'DeviceDetailsComponent',

  components: {
    TerminalDialog,
    DeviceIcon,
    DeviceDelete,
    DeviceRename,
    TagFormUpdate,
  },

  filters: { formatDate, lastSeen, hasPermission },

  data() {
    return {
      uid: '',
      hostname: window.location.hostname,
      hide: true,
      device: null,
      dialogDelete: false,
      dialogError: false,
      deviceRenameShow: false,
      tagFormUpdateShow: false,
      deviceDeleteShow: false,
      renameAction: 'rename',
      updateAction: 'deviceUpdate',
    };
  },

  computed: {
    hasAuthorizationRename() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.device[this.renameAction],
        );
      }

      return false;
    },

    hasAuthorizationFormUpdate() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.tag[this.updateAction],
        );
      }

      return false;
    },
  },

  async created() {
    this.uid = await this.$route.params.id;

    this.getDevice();
  },

  methods: {
    redirect() {
      this.dialogError = false;
      this.$router.push('/devices');
    },

    receiveName(params) {
      this.device.name = params;
    },

    async getDevice() {
      try {
        await this.$store.dispatch('devices/get', this.uid);
        this.device = this.$store.getters['devices/get'];
      } catch (error) {
        this.hide = false;
        this.dialogError = true;
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.deviceDetails);
      }
    },

    openDialog(action) {
      this[action] = !this[action];

      this.closeMenu();
    },

    closeMenu() {
      this.$refs.menu.isActive = false;
    },
  },
};

</script>
