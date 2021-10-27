<template>
  <fragment>
    <v-card-text class="pa-0">
      <v-data-table
        v-model="selected"
        class="elevation-1"
        :headers="headers"
        :items="getListDevices"
        :items-per-page="10"
        :footer-props="{'items-per-page-options': [10, 25, 50, 100]}"
        :server-items-length="getNumberDevices"
        :options.sync="pagination"
        :show-select="disableShowSelect"
        :hide-default-footer="!disableShowSelect"
        item-key="uid"
        data-test="devices-dataTable"
      >
        <template #[`item.hostname`]="{ item }">
          <router-link :to="{ name: 'detailsDevice', params: { id: item.uid } }">
            {{ item.name }}
          </router-link>
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
      </v-data-table>
    </v-card-text>
  </fragment>
</template>
<script>

import DeviceIcon from '@/components/device/DeviceIcon';
import formatDeviceSort from '@/components/filter/object';
import { lastSeen } from '@/components/filter/date';

export default {
  name: 'DeviceListChooser',

  components: {
    DeviceIcon,
  },

  filters: { lastSeen },

  props: {
    action: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      hostname: window.location.hostname,
      pagination: {},
      headers: [
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
      ],
      singleSelect: true,
    };
  },

  computed: {
    getListDevices() {
      return this.$store.getters['devices/getDevicesForUserToChoose'];
    },

    getNumberDevices() {
      return this.$store.getters['devices/getNumberForUserToChoose'];
    },

    selected: {
      get() {
        return this.$store.getters['devices/getDevicesSelected'];
      },

      set(data) {
        this.$store.dispatch('devices/setDevicesSelected', data);
      },
    },

    disableShowSelect() {
      return !(this.getNumberDevices === 3);
    },
  },

  watch: {
    pagination: {
      handler() {
        if (this.action !== 'suggestedDevices') {
          this.getDevices();
        }
      },
      deep: true,
    },
  },

  created() {
    this.$store.dispatch('devices/getDevicesMostUsed');
  },

  methods: {
    async getDevices() {
      const sortStatusMap = formatDeviceSort(
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
        await this.$store.dispatch('devices/setDevicesForUserToChoose', data);
      } catch (error) {
        if (error.response.status === 403) {
          this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.deviceList);
        }
      }
    },

    address(item) {
      return `${item.namespace}.${item.name}@${this.hostname}`;
    },

    copySSHID(sshid) {
      this.$clipboard(sshid);
      this.showCopySnack();
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
