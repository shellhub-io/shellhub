<template>
  <fragment>
    <v-card-text class="pa-0">
      <v-data-table
        class="elevation-1"
        :headers="headers"
        :items="getListDevices"
        data-test="dataTable-field"
        :items-per-page="10"
        :footer-props="{'items-per-page-options': [10, 25, 50, 100]}"
        :server-items-length="getNumberDevices"
        :options.sync="pagination"
      >
        <template #[`item.online`]="{ item }">
          <v-icon
            v-if="item.online"
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
            <span>last seen {{ item.last_seen | lastSeen }}</span>
          </v-tooltip>
        </template>

        <template #[`item.hostname`]="{ item }">
          <router-link :to="{ name: 'detailsDevice', params: { id: item.uid } }">
            {{ item.name }}
          </router-link>
        </template>

        <template #[`item.tags`]="{ item }">
          <div
            v-if="item.tags[0]"
            class="mt-1"
          >
            <v-tooltip
              v-for="(tag, index) in item.tags"
              :key="index"
              bottom
              :disabled="!showTag(tag)"
            >
              <template #activator="{ on, attrs }">
                <v-chip
                  class="ml-1 mb-1"
                  small
                  outlined
                  v-bind="attrs"
                  v-on="on"
                >
                  {{ displayOnlyTenCharacters(tag) }}
                </v-chip>
              </template>

              <span v-if="showTag(tag)">
                {{ tag }}
              </span>
            </v-tooltip>
          </div>
        </template>

        <template #[`item.info.pretty_name`]="{ item }">
          <DeviceIcon
            :icon-name="item.info.id"
            data-test="deviceIcon-component"
          />
          {{ item.info.pretty_name }}
        </template>

        <template #[`item.namespace`]="{ item }">
          <v-chip class="list-itens">
            {{ address(item) }}
            <v-icon
              small
              right
              @click="copySSHID(address(item))"
            >
              mdi-content-copy
            </v-icon>
          </v-chip>
        </template>

        <template #[`item.actions`]="{ item }">
          <v-menu
            :ref="'menu'+getListDevices.indexOf(item)"
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
              <v-list-item @click.stop="detailsDevice(item)">
                <v-icon left>
                  info
                </v-icon>

                <v-list-item-title>
                  Details
                </v-list-item-title>
              </v-list-item>

              <v-list-item @click="showTagDialog(getListDevices.indexOf(item))">
                <TagFormDialog
                  action="create"
                  :uid="item.uid"
                  :show.sync="tagDialogShow[getListDevices.indexOf(item)]"
                  data-test="tagFormDialog-component"
                  @update="getDevices()"
                />
              </v-list-item>

              <v-list-item
                v-if="item.online"
                @click.stop="showTerminalDialog(getListDevices.indexOf(item))"
              >
                <TerminalDialog
                  :uid="item.uid"
                  :show.sync="terminalDialogShow[getListDevices.indexOf(item)]"
                  data-test="terminalDialog-component"
                />
              </v-list-item>

              <v-list-item @click="showDeviceDelete(getListDevices.indexOf(item))">
                <DeviceDelete
                  :uid="item.uid"
                  :show.sync="deviceDeleteShow[getListDevices.indexOf(item)]"
                  data-test="deviceDelete-component"
                  @update="refresh"
                />
              </v-list-item>
            </v-card>
          </v-menu>
        </template>
      </v-data-table>
    </v-card-text>
  </fragment>
</template>
<script>

import TerminalDialog from '@/components/terminal/TerminalDialog';
import DeviceIcon from '@/components/device/DeviceIcon';
import DeviceDelete from '@/components/device/DeviceDelete';
import TagFormDialog from '@/components/setting/tag/TagFormDialog';
import { lastSeen } from '@/components/filter/date';
import formatDeviceSort from '@/components/filter/object';

export default {
  name: 'DeviceListComponent',

  components: {
    TerminalDialog,
    DeviceIcon,
    DeviceDelete,
    TagFormDialog,
  },

  filters: { lastSeen },

  data() {
    return {
      hostname: window.location.hostname,
      pagination: {},
      tags: [],
      tagDialogShow: [],
      terminalDialogShow: [],
      deviceDeleteShow: [],
      headers: [
        {
          text: 'Online',
          value: 'online',
          align: 'center',
        },
        {
          text: 'Hostname',
          value: 'hostname',
          align: 'center',
        },
        {
          text: 'Operating System',
          value: 'info.pretty_name',
          align: 'center',
          sortable: false,
        },
        {
          text: 'SSHID',
          value: 'namespace',
          align: 'center',
          sortable: false,
        },
        {
          text: 'Tags',
          value: 'tags',
          align: 'center',
          sortable: false,
        },
        {
          text: 'Actions',
          value: 'actions',
          align: 'center',
          sortable: false,
        },
      ],
    };
  },

  computed: {
    getListDevices() {
      return this.$store.getters['devices/list'];
    },

    getNumberDevices() {
      return this.$store.getters['devices/getNumberDevices'];
    },
  },

  watch: {
    pagination: {
      handler() {
        this.getDevices();
      },
      deep: true,
    },
  },

  mounted() {
    this.$store.dispatch('devices/resetListDevices');
  },

  methods: {
    async getDevices() {
      let sortStatusMap = {};

      sortStatusMap = formatDeviceSort(
        this.pagination.sortBy[0],
        this.pagination.sortDesc[0],
      );

      const data = {
        perPage: this.pagination.itemsPerPage,
        page: this.pagination.page,
        filter: this.$store.getters['devices/getFilter'],
        status: 'accepted',
        sortStatusField: sortStatusMap.field,
        sortStatusString: sortStatusMap.statusString,
      };

      try {
        await this.$store.dispatch('devices/fetch', data);

        this.setArrays();
      } catch (error) {
        if (error.response.status === 403) {
          this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.deviceList);
        }
      }
    },

    detailsDevice(value) {
      this.$router.push(`/device/${value.uid}`);
    },

    address(item) {
      return `${item.namespace}.${item.name}@${this.hostname}`;
    },

    copySSHID(sshid) {
      this.$clipboard(sshid);
      this.showCopySnack();
    },

    copy(device) {
      this.$clipboard(device.uid);
    },

    showCopySnack() {
      this.$store.dispatch('snackbar/showSnackbarCopy', this.$copy.deviceSSHID);
    },

    refresh() {
      this.getDevices();
    },

    displayOnlyTenCharacters(str) {
      if (str !== undefined) {
        if (str.length > 10) return `${str.substr(0, 10)}...`;
      }
      return str;
    },

    showTag(str) {
      if (str !== undefined) {
        if (str.length > 10) {
          return true;
        }
      }
      return false;
    },

    showTagDialog(index) {
      this.tagDialogShow[index] = this.tagDialogShow[index] === undefined
        ? true : !this.tagDialogShow[index];
      this.$set(this.tagDialogShow, index, this.tagDialogShow[index]);

      this.closeMenu(index);
    },

    showTerminalDialog(index) {
      this.terminalDialogShow[index] = this.terminalDialogShow[index] === undefined
        ? true : !this.terminalDialogShow[index];
      this.$set(this.terminalDialogShow, index, this.terminalDialogShow[index]);

      this.closeMenu(index);
    },

    showDeviceDelete(index) {
      this.deviceDeleteShow[index] = this.deviceDeleteShow[index] === undefined
        ? true : !this.deviceDeleteShow[index];
      this.$set(this.deviceDeleteShow, index, this.deviceDeleteShow[index]);

      this.closeMenu(index);
    },

    setArrays() {
      const numberDevices = this.getListDevices.length;

      if (numberDevices > 0) {
        this.terminalDialogShow = new Array(numberDevices).fill(false);
        this.deviceDeleteShow = new Array(numberDevices).fill(false);
        this.tagDialogShow = new Array(numberDevices).fill(false);
      }
    },

    closeMenu(index) {
      this.$refs[`menu${index}`].isActive = false;
    },
  },
};

</script>

<style scoped>

.list-itens {
  font-family: monospace;
}

.icons{
  margin-right: 4px;
}

.btn-right{
  left: 210px;
}

.short{
  width:140px;
}
.short span{
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

</style>
