<template>
  <fragment>
    <v-card class="mt-2">
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
        data-test="firewallRuleList-dataTable"
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
          <v-menu
            :ref="'menu'+getFirewallRules.indexOf(item)"
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
              <v-list-item @click.stop="showFirewallRuleEdit(getFirewallRules.indexOf(item))">
                <FirewallRuleEdit
                  :firewall-rule="item"
                  :create-rule="false"
                  :show.sync="firewallRuleEditShow[getFirewallRules.indexOf(item)]"
                  data-test="firewallRuleEdit-component"
                  @update="refresh"
                />
              </v-list-item>

              <v-list-item @click.stop="showFirewallRuleDelete(getFirewallRules.indexOf(item))">
                <FirewallRuleDelete
                  :id="item.id"
                  :show.sync="firewallRuleDeleteShow[getFirewallRules.indexOf(item)]"
                  data-test="firewallRuleDelete-component"
                  @update="refresh"
                />
              </v-list-item>
            </v-card>
          </v-menu>
        </template>
      </v-data-table>
    </v-card>
  </fragment>
</template>

<script>

import FirewallRuleEdit from '@/components/firewall_rule/FirewallRuleFormDialog';
import FirewallRuleDelete from '@/components/firewall_rule/FirewallRuleDelete';

export default {
  name: 'FirewallRuleListComponent',

  components: {
    FirewallRuleDelete,
    FirewallRuleEdit,
  },

  data() {
    return {
      pagination: {},
      firewallRuleEditShow: [],
      firewallRuleDeleteShow: [],

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

          this.setArrays();
        } catch (error) {
          if (error.response.status === 403) {
            this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
          } else {
            this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.firewallRuleList);
          }
        }
      } else {
        this.setArrays();
        this.$store.dispatch('boxs/setStatus', false);
      }
    },

    showFirewallRuleEdit(index) {
      this.firewallRuleEditShow[index] = this.firewallRuleEditShow[index] === undefined
        ? true : !this.firewallRuleEditShow[index];
      this.$set(this.firewallRuleEditShow, index, this.firewallRuleEditShow[index]);

      this.closeMenu(index);
    },

    showFirewallRuleDelete(index) {
      this.firewallRuleDeleteShow[index] = this.firewallRuleDeleteShow[index] === undefined
        ? true : !this.firewallRuleDeleteShow[index];
      this.$set(this.firewallRuleDeleteShow, index, this.firewallRuleDeleteShow[index]);

      this.closeMenu(index);
    },

    setArrays() {
      const numberFirewallRules = this.getFirewallRules.length;

      if (numberFirewallRules > 0) {
        this.firewallRuleEditShow = new Array(numberFirewallRules).fill(false);
        this.firewallRuleDeleteShow = new Array(numberFirewallRules).fill(false);
      }
    },

    closeMenu(index) {
      this.$refs[`menu${index}`].isActive = false;
    },
  },
};
</script>
