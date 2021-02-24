<template>
  <fragment>
    <div class="d-flex pa-0 align-center">
      <h1>Devices</h1>
      <v-spacer />
      <v-text-field
        v-model="search"
        append-icon="mdi-magnify"
        label="Search by hostname"
        class="mx-6"
        single-line
        hide-details
        data-test="search-text"
      />
      <v-spacer />
      <DeviceAdd />
    </div>
    <v-card class="mt-2">
      <v-app-bar
        flat
        color="transparent"
      >
        <v-tabs>
          <v-tab
            to="/devices"
          >
            Device List
          </v-tab>
          <v-tab
            to="/devices/pending"
          >
            <v-badge
              :content="getNumberPendingDevices"
              :value="getNumberPendingDevices"
              data-test="badge-field"
              overlap
              inline
              color="success"
            >
              Pending
            </v-badge>
          </v-tab>
          <v-tab
            to="/devices/rejected"
          >
            Rejected
          </v-tab>
        </v-tabs>
      </v-app-bar>
      <v-divider />

      <router-view />
    </v-card>
  </fragment>
</template>

<script>

import DeviceAdd from '@/components/device/DeviceAdd';

export default {
  name: 'DeviceList',

  components: {
    DeviceAdd,
  },

  data() {
    return {
      search: '',
    };
  },

  computed: {
    getNumberPendingDevices() {
      return this.$store.getters['stats/stats'].pending_devices;
    },
  },

  watch: {
    search() {
      this.getDevices();
    },
  },

  async created() {
    try {
      await this.$store.dispatch('stats/get');
    } catch (e) {
      if (e.response.status === 403) {
        this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
      } else {
        this.$store.dispatch('snackbar/showSnackbarErrorDefault');
      }
    }
  },

  async destroyed() {
    await this.$store.dispatch('devices/setFilter', null);
  },

  methods: {
    async getDevices() {
      let encodedFilter = null;

      if (this.search) {
        const filter = [{ type: 'property', params: { name: 'name', operator: 'like', value: this.search } }];
        encodedFilter = btoa(JSON.stringify(filter));
      }
      await this.$store.dispatch('devices/setFilter', encodedFilter);

      try {
        this.$store.dispatch('devices/refresh');
      } catch (e) {
        if (e.response.status === 403) {
          this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorDefault');
        }
      }
    },

    formatSortObject(field, isDesc) {
      let formatedField = null;
      let formatedStatus = false;
      let ascOrDesc = 'asc';

      if (field !== undefined) {
        formatedField = field === 'hostname' ? 'name' : field; // customize to api field
      }

      if (isDesc !== undefined) {
        formatedStatus = isDesc;
      }

      if (formatedStatus === true) {
        ascOrDesc = 'desc';
      }

      return {
        field: formatedField,
        status: formatedStatus,
        statusString: ascOrDesc,
      };
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
