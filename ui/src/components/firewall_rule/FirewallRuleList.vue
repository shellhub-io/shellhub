<template>
  <fragment>
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
          :items="getFirewallRules"
          data-test="dataTable-field"
          item-key="uid"
          :sort-by="['started_at']"
          :sort-desc="[true]"
          :items-per-page="10"
          :footer-props="{'items-per-page-options': [10, 25, 50, 100]}"
          :server-items-length="getNumberFirewallRules"
          :options.sync="pagination"
          :disable-sort="true"
        >
          <template #[`item.active`]="{ item }">
            <v-icon
              v-if="item.active"
              color="success"
            >
              check_circle
            </v-icon>
            <v-icon
              v-else
              bottom
            >
              check_circle
            </v-icon>
          </template>

          <template #[`item.priority`]="{ item }">
            {{ item.priority }}
          </template>

          <template #[`item.action`]="{ item }">
            {{ item.action }}
          </template>

          <template #[`item.source_ip`]="{ item }">
            {{ item.source_ip }}
          </template>

          <template #[`item.username`]="{ item }">
            {{ item.username }}
          </template>

          <template #[`item.hostname`]="{ item }">
            {{ item.hostname }}
          </template>

          <template #[`item.actions`]="{ item }">
            <FirewallRuleEdit
              v-if="isOwner"
              data-test="firewall-dialog-field-2"
              :firewall-rule="item"
              :create-rule="false"
              @update="refresh"
            />
            <FirewallRuleDelete
              v-if="isOwner"
              :id="item.id"
              data-test="firewall-delete-field"
              @update="refresh"
            />
          </template>
        </v-data-table>
      </v-card-text>
    </v-card>
  </fragment>
</template>

<script>

import FirewallRuleEdit from '@/components/firewall_rule/FirewallRuleFormDialog';
import FirewallRuleDelete from '@/components/firewall_rule/FirewallRuleDelete';

export default {
  name: 'FirewallRuleList',

  components: {
    FirewallRuleDelete,
    FirewallRuleEdit,
  },

  data() {
    return {
      pagination: {},

      headers: [
        {
          text: 'Active',
          value: 'active',
          align: 'center',
        },
        {
          text: 'Priority',
          value: 'priority',
          align: 'center',
        },
        {
          text: 'Action',
          value: 'action',
          align: 'center',
        },
        {
          text: 'Source IP',
          value: 'source_ip',
          align: 'center',
        },
        {
          text: 'Username',
          value: 'username',
          align: 'center',
        },
        {
          text: 'Hostname',
          value: 'hostname',
          align: 'center',
        },
        {
          text: 'Actions',
          value: 'actions',
          align: 'center',
        },
      ],

      showHelp: false,
    };
  },

  computed: {
    getFirewallRules() {
      return this.$store.getters['firewallrules/list'];
    },

    getNumberFirewallRules() {
      return this.$store.getters['firewallrules/getNumberFirewalls'];
    },

    isOwner() {
      return this.$store.getters['namespaces/owner'];
    },
  },

  watch: {
    pagination: {
      handler() {
        this.getFirewalls();
      },
      deep: true,
    },
  },

  methods: {
    refresh() {
      this.getFirewalls();
    },

    async getFirewalls() {
      if (!this.$store.getters['boxs/getStatus']) {
        const data = {
          perPage: this.pagination.itemsPerPage,
          page: this.pagination.page,
        };

        try {
          await this.$store.dispatch('firewallrules/fetch', data);
        } catch (e) {
          if (e.response.status === 403) {
            this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
          } else {
            this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.firewallRuleList);
          }
        }
      } else {
        this.$store.dispatch('boxs/setStatus', false);
      }
    },
  },
};
</script>
