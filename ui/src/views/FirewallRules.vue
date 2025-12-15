<template>
  <PageHeader
    icon="mdi-security"
    title="Firewall Rules"
    overline="Security"
    icon-color="primary"
    data-test="firewall-rules"
  >
    <template #description>
      <p class="mb-0">
        Control which SSH connections reach your devices with fine-grained firewall rules.
        Allow or deny connections from specific IP addresses to devices using specific usernames.
      </p>
    </template>
    <template #actions>
      <FirewallRuleAdd @update="fetchFirewallRules" />
    </template>
  </PageHeader>

  <div>
    <FirewallRuleList v-if="hasFirewallRule" />

    <NoItemsMessage
      v-else
      item="Firewall Rules"
      icon="mdi-security"
      data-test="no-items-message-component"
    >
      <template #content>
        <p>
          ShellHub provides flexible firewall for filtering SSH connections.
          It gives a fine-grained control over which SSH connections reach the devices.
        </p>
        <p>
          Using Firewall Rules you can deny or allow SSH connections from specific
          IP addresses to a specific or a group of devices using a given username.
        </p>
      </template>
      <template #action>
        <FirewallRuleAdd @update="fetchFirewallRules" />
      </template>
    </NoItemsMessage>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import NoItemsMessage from "../components/NoItemsMessage.vue";
import { envVariables } from "../envVariables";
import FirewallRuleList from "../components/firewall/FirewallRuleList.vue";
import FirewallRuleAdd from "../components/firewall/FirewallRuleAdd.vue";
import PageHeader from "../components/PageHeader.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useFirewallRulesStore from "@/store/modules/firewall_rules";
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
</script>
