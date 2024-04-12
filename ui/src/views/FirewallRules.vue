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
      href="https://docs.shellhub.io/user-guides/security/managing-firewall-rules"
      target="_blank"
      rel="noopener noreferrer"
    >See More</a
    >
  </p>

  <div>
    <FirewallRuleList v-if="hasFirewallRule" />

    <BoxMessage
      v-if="showBoxMessage"
      typeMessage="firewall"
      data-test="BoxMessageFirewall-component"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useStore } from "../store";
import BoxMessage from "../components/Box/BoxMessage.vue";
import FirewallRuleList from "../components/firewall/FirewallRuleList.vue";
import FirewallRuleAdd from "../components/firewall/FirewallRuleAdd.vue";
import { INotificationsError } from "../interfaces/INotifications";
import handleError from "@/utils/handleError";

const showHelp = ref(false);
const store = useStore();
const show = ref(false);
const hasFirewallRule = computed(
  () => store.getters["firewallRules/getNumberFirewalls"] > 0,
);
const showBoxMessage = computed(() => !hasFirewallRule.value && show.value);

const refresh = async () => {
  try {
    await store.dispatch("firewallRules/refresh");
  } catch (error: unknown) {
    store.dispatch(
      "snackbar/showSnackbarErrorLoading",
      INotificationsError.firewallRuleList,
    );
    handleError(error);
  }
};

onMounted(async () => {
  try {
    store.dispatch("box/setStatus", true);
    store.dispatch("firewallRules/resetPagePerpage");
    await refresh();
    show.value = true;
  } catch (error: unknown) {
    handleError(error);
  }
});

defineExpose({ showHelp, show });
</script>
