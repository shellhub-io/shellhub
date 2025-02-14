<template>
  <div>
    <h1>Firewall Rules</h1>
  </div>
  <v-card class="mt-2">
    <FirewallRulesList />
  </v-card>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import FirewallRulesList from "../components/FirewallRules/FirewallRulesList.vue";

export default defineComponent({
  components: { FirewallRulesList },
});
</script>
