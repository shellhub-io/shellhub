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

    <FirewallRuleAdd @update="refresh" />
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
        <FirewallRuleAdd @update="refresh" />
      </template>
    </NoItemsMessage>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useStore } from "../store";
import NoItemsMessage from "../components/NoItemsMessage.vue";
import { envVariables } from "../envVariables";
import FirewallRuleList from "../components/firewall/FirewallRuleList.vue";
import FirewallRuleAdd from "../components/firewall/FirewallRuleAdd.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";

const showHelp = ref(false);
const store = useStore();
const snackbar = useSnackbar();
const hasFirewallRule = computed(
  () => store.getters["firewallRules/getNumberFirewalls"] > 0,
);

const refresh = async () => {
  try {
    await store.dispatch("firewallRules/refresh");
  } catch (error: unknown) {
    snackbar.showError("Failed to load the firewall rules list.");
    handleError(error);
  }
};

onMounted(async () => {
  try {
    store.dispatch("firewallRules/resetPagePerpage");
    if (!envVariables.isCommunity) {
      await refresh();
    }
  } catch (error: unknown) {
    handleError(error);
  }
});

defineExpose({ showHelp });
</script>
