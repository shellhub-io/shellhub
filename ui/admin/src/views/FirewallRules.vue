<template>
  <div>
    <h1>Firewall Rules</h1>
  </div>
  <v-card class="mt-2">
    <FirewallRulesList />
  </v-card>
</template>

<script setup lang="ts">
import FirewallRulesList from "../components/FirewallRules/FirewallRulesList.vue";
</script>
