<template>
  <fragment>
    <div class="d-flex pa-0 align-center">
      <h1>Firewall Rules</h1>
      <v-btn
        icon
        x-small
        class="ml-2"
        @click="showHelp = !showHelp"
      >
        <v-icon>mdi-help-circle</v-icon>
      </v-btn>
      <v-spacer />
      <v-spacer />
      <FirewallRuleFormDialog
        v-if="isOwner"
        :create-rule="true"
        @update="refresh"
      />
    </div>
    <p v-if="showHelp">
      Firewall rules gives a fine-grained control over which SSH connections reach the devices.
      <a
        target="_blank"
        href="https://docs.shellhub.io/user-manual/managing-firewall-rules/"
      >See More</a>
    </p>

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
            <FirewallRuleFormDialog
              v-if="isOwner"
              :firewall-rule="item"
              :create-rule="false"
              @update="refresh"
            />
            <FirewallDelete
              v-if="isOwner"
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
    },
  },
};
</script>
