<template>
  <fragment>
    <div class="d-flex pa-0 align-center">
      <h1>Public Keys</h1>
      <v-spacer />
      <v-spacer />
      <PublicKeyFormDialog
        :create-public-key="true"
        @update="refresh"
      />
    </div>

    <v-card class="mt-2">
      <v-app-bar
        flat
        color="transparent"
      >
        <v-toolbar-title />
      </v-app-bar>

      <v-divider />

      <v-card-text class="pa-0">
        <v-data-table
          :headers="headers"
          :items="getPublicKeys"
          data-test="dataTable-field"
          item-key="fingerprint"
          :sort-by="['started_at']"
          :sort-desc="[true]"
          :items-per-page="10"
          :footer-props="{'items-per-page-options': [10, 25, 50, 100]}"
          :server-items-length="getNumberPublicKeys"
          :options.sync="pagination"
          :disable-sort="true"
        >
          <template #[`item.name`]="{ item }">
            {{ item.name }}
          </template>

          <template #[`item.fingerprint`]="{ item }">
            {{ item.fingerprint }}
          </template>

          <template #[`item.created_at`]="{ item }">
            {{ item.created_at | moment("ddd, MMM Do YY, h:mm:ss a") }}
          </template>

          <template #[`item.actions`]="{ item }">
            <PublicKeyFormDialog
              :public-key="item"
              :create-public-key="false"
              @update="refresh"
            />
            <PublicKeyDelete
              :fingerprint="item.fingerprint"
              @update="refresh"
            />
          </template>
        </v-data-table>
      </v-card-text>
    </v-card>
  </fragment>
</template>

<script>

import PublicKeyFormDialog from '@/components/public_key/PublicKeyFormDialog';
import PublicKeyDelete from '@/components/public_key/PublicKeyDelete';

export default {
  name: 'PublickeyList',

  components: {
    PublicKeyFormDialog,
    PublicKeyDelete,
  },

  data() {
    return {
      pagination: {},

      headers: [
        {
          text: 'Name',
          value: 'name',
          align: 'center',
        },
        {
          text: 'Fingerprint',
          value: 'fingerprint',
          align: 'center',
        },
        {
          text: 'Created At',
          value: 'created_at',
          align: 'center',
        },
        {
          text: 'Actions',
          value: 'actions',
          align: 'center',
        },
      ],
    };
  },

  computed: {
    getPublicKeys() {
      return this.$store.getters['publickeys/list'];
    },

    getNumberPublicKeys() {
      return this.$store.getters['publicKeys/getNumberPublicKeys'];
    },
  },

  watch: {
    pagination: {
      handler() {
        this.getPublicKeysList();
      },
      deep: true,
    },
  },

  methods: {
    refresh() {
      this.getPublicKeysList();
    },

    async getPublicKeysList() {
      const data = {
        perPage: this.pagination.itemsPerPage,
        page: this.pagination.page,
      };

      try {
        await this.$store.dispatch('publickeys/fetch', data);
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.publicKeyList);
      }
    },
  },
};

</script>
