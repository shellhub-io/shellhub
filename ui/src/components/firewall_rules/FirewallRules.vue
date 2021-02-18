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
        data-test="firewall-dialog-field"
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
      <router-view v-if="hasFirewallRules" />
      <BoxMessageFirewall
        v-else
        type-message="firewall"
      />
    </v-card>
  </fragment>
</template>

<script>

import FirewallRuleFormDialog from '@/components/firewall_rules/FirewallRulesFormDialog';
import BoxMessageFirewall from '@/components/box/BoxMessage';

export default {
  name: 'Firewall',

  components: {
    FirewallRuleFormDialog,
    BoxMessageFirewall,
  },

  data() {
    return {
      showHelp: false,
    };
  },

  computed: {
    isOwner() {
      return this.$store.getters['namespaces/owner'];
    },

    hasFirewallRules() {
      return this.$store.getters['firewallrules/getNumberFirewalls'] > 0;
    },
  },

  async created() {
    await this.refresh();
  },

  methods: {
    async refresh() {
      try {
        await this.$store.dispatch('firewallrules/refresh');
      } catch (e) {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.firewallRuleList);
      }
    },
  },
};
</script>
