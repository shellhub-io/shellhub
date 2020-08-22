<template>
  <fragment>
    <div class="d-flex pa-0 align-center">
      <h1>Firewall Rules</h1>
      <v-spacer />
      <v-spacer />
      <FirewallRuleFormDialog
        :create-rule="true"
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
          :items="getFirewallRules"
          item-key="uid"
          :sort-by="['started_at']"
          :sort-desc="[true]"
          :items-per-page="10"
          :footer-props="{'items-per-page-options': [10, 25, 50, 100]}"
          :server-items-length="getNumberFirewallRules"
          :options.sync="pagination"
          :disable-sort="true"
        >
          <template v-slot:item.active="{ item }">
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

          <template v-slot:item.priority="{ item }">
            {{ item.priority }}
          </template>

          <template v-slot:item.action="{ item }">
            {{ item.action }}
          </template>

          <template v-slot:item.source_ip="{ item }">
            {{ item.source_ip }}
          </template>

          <template v-slot:item.username="{ item }">
            {{ item.username }}
          </template>

          <template v-slot:item.hostname="{ item }">
            {{ item.hostname }}
          </template>

          <template v-slot:item.actions="{ item }">
            <FirewallRuleFormDialog
              :firewall-rule="item"
              :create-rule="false"
              @update="refresh"
            />
            <FirewallDelete
              :id="item.id"
              @update="refresh"
            />
          </template>
        </v-data-table>
      </v-card-text>
    </v-card>
  </fragment>
</template>

<script>

import FirewallRuleFormDialog from '@/components/firewall_rules/FirewallRulesFormDialog';
import FirewallDelete from '@/components/firewall_rules/FirewallRulesDelete';

export default {
  name: 'FirewallList',

  components: {
    FirewallDelete,
    FirewallRuleFormDialog,
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
    };
  },

  computed: {
    getFirewallRules() {
      return this.$store.getters['firewallrules/list'];
    },

    getNumberFirewallRules() {
      return this.$store.getters['firewallrules/getNumberFirewalls'];
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
      const data = {
        perPage: this.pagination.itemsPerPage,
        page: this.pagination.page,
      };

      try {
        await this.$store.dispatch('firewallrules/fetch', data);
      } catch {
        this.$store.dispatch('modals/showSnackbarError', true);
      }
    },
  },
};
</script>
