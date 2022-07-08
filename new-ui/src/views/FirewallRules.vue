<template>
  <div
    class="d-flex flex-column justify-space-between align-center flex-sm-row mb-2"
  >
    <div class="d-flex align-center">
      <h1>Firewall Rules</h1>

      <v-icon @click="showHelp = !showHelp" class="ml-2" size="small">
        mdi-help-circle
      </v-icon>

      <v-spacer />
      <v-spacer />
    </div>

    <FirewallRuleAdd @update="refresh"/>
  </div>

  <p v-if="showHelp" class="mt-n4 mb-2">
    Firewall rules gives a fine-grained control over which SSH connections reach
    the devices.
    <a
      target="_blank"
      href="https://docs.shellhub.io/user-manual/managing-firewall-rules/"
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

<script lang="ts">
import { computed, defineComponent, onMounted, ref } from "vue";
import { useStore } from "../store";
import BoxMessage from "../components/Box/BoxMessage.vue";
import FirewallRuleList from "../components/firewall/FirewallRuleList.vue";
import FirewallRuleAdd from "../components/firewall/FirewallRuleAdd.vue";
import { INotificationsError } from "../interfaces/INotifications";

export default defineComponent({
  setup() {
    const showHelp = ref(false);
    const store = useStore();
    const show = ref(false);
    const hasFirewallRule = computed(
      () => store.getters["firewallRules/getNumberFirewalls"] > 0
    );
    const showBoxMessage = computed(() => !hasFirewallRule.value && show.value);
    
    onMounted(async () => {
      store.dispatch("box/setStatus", true);
      store.dispatch("firewallRules/resetPagePerpage");
      await refresh();
      store.dispatch("tags/fetch");
      show.value = true;
    });

    const refresh = async () => {
      try {
        await store.dispatch("firewallRules/refresh");
      } catch {
        store.dispatch(
          "snackbar/showSnackbarErrorLoading",
          INotificationsError.firewallRuleList
        );
      }
    };

    return {
      show,
      showHelp,
      hasFirewallRule,
      showBoxMessage,
      refresh,
    };
  },
  components: { BoxMessage, FirewallRuleList, FirewallRuleAdd },
});
</script>
