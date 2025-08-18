<template>
  <div
    class="d-flex flex-column justify-space-between align-center flex-sm-row mb-2"
    data-test="firewall-rules"
  >
    <div class="d-flex align-center">
      <h1>Firewall Rules</h1>

      <v-icon @click="showHelp = !showHelp" class="ml-2" size="small" data-test="help-icon">
        mdi-help-circle
      </v-icon>

      <v-spacer />
      <v-spacer />
    </div>

    <FirewallRuleAdd @update="fetchFirewallRules" />
  </div>

  <p v-if="showHelp" class="mt-n4 mb-2" data-test="firewall-helper">
    Firewall rules gives a fine-grained control over which SSH connections reach
    the devices.
    <a
      href="https://docs.shellhub.io/user-guides/firewall/"
      target="_blank"
      rel="noopener noreferrer"
    >See More</a
    >
  </p>

  <div>
    <FirewallRuleList v-if="hasFirewallRule" />

    <NoItemsMessage
      v-else
      item="Firewall Rules"
      icon="mdi-security"
      data-test="no-items-message-component"
    >
      <template #content>
        <p>ShellHub provides flexible firewall for filtering SSH connections.
          It gives a fine-grained control over which SSH connections reach the devices.</p>
        <p>Using Firewall Rules you can deny or allow SSH connections from specific
          IP addresses to a specific or a group of devices using a given username.</p>
      </template>
      <template #action>
        <FirewallRuleAdd @update="fetchFirewallRules" />
      </template>
    </NoItemsMessage>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import NoItemsMessage from "../components/NoItemsMessage.vue";
import { envVariables } from "../envVariables";
import FirewallRuleList from "../components/firewall/FirewallRuleList.vue";
import FirewallRuleAdd from "../components/firewall/FirewallRuleAdd.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useFirewallRulesStore from "@/store/modules/firewall_rules";

const showHelp = ref(false);
const firewallRulesStore = useFirewallRulesStore();
const snackbar = useSnackbar();
const hasFirewallRule = computed(() => firewallRulesStore.firewallRuleCount > 0);

const fetchFirewallRules = async () => {
  try {
    await firewallRulesStore.fetchFirewallRuleList();
  } catch (error: unknown) {
    snackbar.showError("Failed to load the firewall rules list.");
    handleError(error);
  }
};

onMounted(async () => { if (!envVariables.isCommunity) await fetchFirewallRules(); });

defineExpose({ showHelp });
</script>
