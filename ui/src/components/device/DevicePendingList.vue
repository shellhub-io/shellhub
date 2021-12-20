<template>
  <fragment>
    <v-card-text class="pa-0">
      <v-data-table
        class="elevation-1"
        :headers="headers"
        :items="getListPendingDevices"
        data-test="dataTable-field"
        :items-per-page="10"
        :footer-props="{'items-per-page-options': [10, 25, 50, 100]}"
        :server-items-length="getNumberPendingDevices"
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
          <DeviceIcon
            :icon-name="item.info.id"
            data-test="deviceIcon-component"
          />
          {{ item.info.pretty_name }}
        </template>

        <template #[`item.request_time`]="{ item }">
          {{ [item.last_seen] | moment("ddd, MMM Do YY, h:mm:ss a") }}
        </template>

        <template #[`item.actions`]="{ item }">
          <v-menu
            :ref="'menu'+getListPendingDevices.indexOf(item)"
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
              <v-list-item
                @click="showDeviceAcceptButton(getListPendingDevices.indexOf(item))"
              >
                <DeviceActionButton
                  :uid="item.uid"
                  action="accept"
                  :show.sync="deviceAcceptButtonShow[getListPendingDevices.indexOf(item)]"
                  data-test="DeviceActionButtonAccept-component"
                  @update="refresh"
                />
              </v-list-item>

              <v-list-item
                @click="showDeviceRejectButton(getListPendingDevices.indexOf(item))"
              >
                <DeviceActionButton
                  :uid="item.uid"
                  action="reject"
                  :show.sync="deviceRejectButtonShow[getListPendingDevices.indexOf(item)]"
                  data-test="deviceActionButtonReject-component"
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

import DeviceIcon from '@/components/device/DeviceIcon';
import DeviceActionButton from '@/components/device/DeviceActionButton';
import formatDeviceSort from '@/components/filter/object';

export default {
  name: 'DeviceListComponent',

  components: {
    DeviceIcon,
    DeviceActionButton,
  },

  data() {
    return {
      pagination: {},
      deviceAcceptButtonShow: [],
      deviceRejectButtonShow: [],

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

  mounted() {
    this.$store.dispatch('devices/resetListDevices');
  },

  methods: {
    async getPendingDevices() {
      let sortStatusMap = {};

      sortStatusMap = formatDeviceSort(
        this.pagination.sortBy[0],
        this.pagination.sortDesc[0],
      );

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

        this.setArrays();
      } catch (error) {
        if (error.response.status === 403) {
          this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.deviceListPending);
        }
      }
    },

    refresh() {
      this.getPendingDevices();
    },

    showDeviceAcceptButton(index) {
      this.deviceAcceptButtonShow[index] = this.deviceAcceptButtonShow[index] === undefined
        ? true : !this.deviceAcceptButtonShow[index];
      this.$set(this.deviceAcceptButtonShow, index, this.deviceAcceptButtonShow[index]);

      this.closeMenu(index);
    },

    showDeviceRejectButton(index) {
      this.deviceRejectButtonShow[index] = this.deviceRejectButtonShow[index] === undefined
        ? true : !this.deviceRejectButtonShow[index];
      this.$set(this.deviceRejectButtonShow, index, this.deviceRejectButtonShow[index]);

      this.closeMenu(index);
    },

    setArrays() {
      const numberPedingDevices = this.getListPendingDevices.length;

      if (numberPedingDevices > 0) {
        this.deviceAcceptButtonShow = new Array(numberPedingDevices).fill(false);
        this.deviceRejectButtonShow = new Array(numberPedingDevices).fill(false);
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

</style>
