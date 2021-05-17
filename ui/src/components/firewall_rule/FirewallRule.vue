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

      <FirewallRuleCreate
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
      <router-view v-if="hasFirewallRule" />

      <BoxMessageFirewall
        v-if="showBoxMessage"
        type-message="firewall"
      />
    </v-card>
  </fragment>
</template>

<script>

import FirewallRuleCreate from '@/components/firewall_rule/FirewallRuleFormDialog';
import BoxMessageFirewall from '@/components/box/BoxMessage';

export default {
  name: 'FirewallRule',

  components: {
    FirewallRuleCreate,
    BoxMessageFirewall,
  },

  data() {
    return {
      showHelp: false,
      show: false,
    };
  },

  computed: {
    hasFirewallRule() {
      return this.$store.getters['firewallrules/getNumberFirewalls'] > 0;
    },

    showBoxMessage() {
      return !this.hasFirewallRule && this.show;
    },
  },

  async created() {
    this.$store.dispatch('boxs/setStatus', true);
    this.$store.dispatch('firewallrules/resetPagePerpage');

    await this.refresh();
    this.show = true;
  },

  methods: {
    async refresh() {
      try {
        await this.$store.dispatch('firewallrules/refresh');
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.firewallRuleList);
      }
    },
  },
};
</script>
