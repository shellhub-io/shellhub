<template>
  <fragment>
    <v-data-table
      :headers="headers"
      :items="getListTokens"
      :server-items-length="getNumberTokens"
      hide-default-footer
      data-test="dataTable-field"
    >
      <template #[`item.tenant_id`]="{ item }">
        {{ item.tenant_id }}
      </template>

      <template #[`item.read_only`]="{ item }">
        {{ item.read_only }}
      </template>

      <template #[`item.actions`]="{ item }">
        <TokenEdit
          :token="item"
          data-test="token-edit-field"
          @update="refresh"
        />

        <TokenDelete
          :id="item.id"
          data-test="token-delete-field"
          @update="refresh"
        />
      </template>
    </v-data-table>
  </fragment>
</template>

<script>

import TokenEdit from '@/components/setting/token/TokenEdit';
import TokenDelete from '@/components/setting/token/TokenDelete';

export default {
  name: 'SettingToken',

  components: {
    TokenEdit,
    TokenDelete,
  },

  data() {
    return {
      pagination: {},
      dialog: true,
      privatekeyPrivacyPolicy: false,
      headers: [
        {
          text: 'Tenant ID',
          value: 'tenant_id',
          align: 'center',
        },
        {
          text: 'Permission',
          value: 'read_only',
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
    getListTokens() {
      return this.$store.getters['tokens/list'];
    },

    getNumberTokens() {
      return this.$store.getters['tokens/getNumberTokens'];
    },
  },

  methods: {
    async refresh() {
      try {
        await this.$store.dispatch('tokens/fetch');
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.tokenList);
      }
    },
  },
};
</script>
