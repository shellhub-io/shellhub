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

        <template #[`item.info.pretty_name`]="{ item }">
          <DeviceIcon :icon-name="item.info.id" />
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
          <v-tooltip bottom>
            <template #activator="{ on }">
              <v-icon
                class="icons"
                v-on="on"
                @click="detailsDevice(item)"
              >
                info
              </v-icon>
            </template>
            <span>Details</span>
          </v-tooltip>

          <TerminalDialog
            v-if="item.online"
            :uid="item.uid"
          />

          <DeviceDelete
            v-if="isOwner"
            :uid="item.uid"
            @update="refresh"
          />
        </template>
      </v-data-table>
    </v-card-text>
  </fragment>
</template>
<script>

import TerminalDialog from '@/components/terminal/TerminalDialog';
import DeviceIcon from '@/components/device/DeviceIcon';
import DeviceDelete from '@/components/device/DeviceDelete';
import formatOrdering from '@/components/device/Device';
import { lastSeen } from '@/components/filter/date';

export default {
  name: 'DeviceList',

  components: {
    TerminalDialog,
    DeviceIcon,
    DeviceDelete,
  },

  filters: { lastSeen },

  mixins: [formatOrdering],

  data() {
    return {
      hostname: window.location.hostname,
      pagination: {},
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

    isOwner() {
      return this.$store.getters['namespaces/owner'];
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

      sortStatusMap = this.formatSortObject(this.pagination.sortBy[0], this.pagination.sortDesc[0]);

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
      } catch (e) {
        if (e.response.status === 403) {
          this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.deviceList);
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

</style>
