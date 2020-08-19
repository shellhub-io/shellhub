<template>
  <fragment>
    <v-card-text class="pa-0">
      <v-data-table
        class="elevation-1"
        :headers="headers"
        :items="getListPendingDevices"
        :items-per-page="10"
        :footer-props="{'items-per-page-options': [10, 25, 50, 100]}"
        :server-items-length="getNumberPendingDevices"
        :options.sync="pagination"
      >
        <template slot="no-data">
          There are no more pending devices
        </template>

        <template v-slot:item.hostname="{ item }">
          <router-link :to="{ name: 'detailsDevice', params: { id: item.uid } }">
            {{ item.name }}
          </router-link>
        </template>

        <template v-slot:item.info.pretty_name="{ item }">
          <DeviceIcon :icon-name="item.info.id" />
          {{ item.info.pretty_name }}
        </template>

        <template v-slot:item.request_time="{ item }">
          {{ item.last_seen | moment("ddd, MMM Do YY, h:mm:ss a") }}
        </template>

        <template v-slot:item.actions="{ item }">
          <DeviceActionButton
            :uid="item.uid"
            action="accept"
            @update="refresh"
          />

          <DeviceActionButton
            :uid="item.uid"
            action="reject"
            @update="refresh"
          />
        </template>
      </v-data-table>
    </v-card-text>
  </fragment>
</template>

<script>

import DeviceIcon from '@/components/device/DeviceIcon';
import DeviceActionButton from '@/components/device/DeviceActionButton';
import formatOrdering from '@/components/device/Device';

export default {
  name: 'DeviceList',

  components: {
    DeviceIcon,
    DeviceActionButton,
  },

  mixins: [formatOrdering],

  data() {
    return {
      pagination: {},
      copySnack: false,
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
          text: 'Request Time',
          value: 'request_time',
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
    getListPendingDevices() {
      return this.$store.getters['devices/list'];
    },

    getNumberPendingDevices() {
      return this.$store.getters['devices/getNumberDevices'];
    },
  },

  watch: {
    pagination: {
      handler() {
        this.getPendingDevices();
      },
      deep: true,
    },
  },

  methods: {
    async getPendingDevices() {
      let sortStatusMap = {};

      sortStatusMap = this.formatSortObject(this.pagination.sortBy[0], this.pagination.sortDesc[0]);

      const data = {
        perPage: this.pagination.itemsPerPage,
        page: this.pagination.page,
        filter: this.$store.getters['devices/getFilter'],
        status: 'pending',
        sortStatusField: sortStatusMap.field,
        sortStatusString: sortStatusMap.statusString,
      };

      try {
        await this.$store.dispatch('devices/fetch', data);
      } catch {
        this.$store.dispatch('modals/showSnackbarError', true);
      }
    },

    refresh() {
      this.getPendingDevices();
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
