<template>
  <fragment>
    <v-card-text class="pa-0">
      <v-data-table
        class="elevation-1"
        :headers="headers"
        :items="getListRejectedDevices"
        data-test="dataTable-field"
        :items-per-page="10"
        :footer-props="{'items-per-page-options': [10, 25, 50, 100]}"
        :server-items-length="getNumberRejectedDevices"
        :options.sync="pagination"
      >
        <template slot="no-data">
          There are no more pending devices
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

        <template #[`item.request_time`]="{ item }">
          {{ item.last_seen | moment("ddd, MMM Do YY, h:mm:ss a") }}
        </template>

        <template #[`item.actions`]="{ item }">
          <DeviceActionButton
            v-if="isOwner"
            :uid="item.uid"
            action="accept"
            @update="refresh"
          />

          <DeviceActionButton
            v-if="isOwner"
            :uid="item.uid"
            action="remove"
            @update="refresh"
          />
        </template>
      </v-data-table>
    </v-card-text>
  </fragment>
</template>

<script>

import DeviceIcon from '@/components/device//DeviceIcon';
import DeviceActionButton from '@/components/device/DeviceActionButton';
import formatOrdering from '@/components/device//Device';

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
    getListRejectedDevices() {
      return this.$store.getters['devices/list'];
    },

    getNumberRejectedDevices() {
      return this.$store.getters['devices/getNumberDevices'];
    },

    isOwner() {
      return this.$store.getters['namespaces/owner'];
    },
  },

  watch: {
    pagination: {
      handler() {
        this.getRejectedDevices();
      },
      deep: true,
    },
  },

  mounted() {
    this.$store.dispatch('devices/resetListDevices');
  },

  methods: {
    async getRejectedDevices() {
      let sortStatusMap = {};

      sortStatusMap = this.formatSortObject(this.pagination.sortBy[0], this.pagination.sortDesc[0]);

      const data = {
        perPage: this.pagination.itemsPerPage,
        page: this.pagination.page,
        filter: this.$store.getters['devices/getFilter'],
        status: 'rejected',
        sortStatusField: sortStatusMap.field,
        sortStatusString: sortStatusMap.statusString,
      };

      try {
        await this.$store.dispatch('devices/fetch', data);
      } catch (e) {
        if (e.response.status === 403) {
          this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.deviceListRejected);
        }
      }
    },

    refresh() {
      this.getRejectedDevices();
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
