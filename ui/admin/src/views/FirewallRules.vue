<template>
  <PageHeader
    icon="mdi-security"
    title="Firewall Rules"
    overline="Security Controls"
    description="Review every policy applied across namespaces and confirm access is locked down."
    icon-color="primary"
    data-test="firewall-rules"
  />
  <FirewallRulesList data-test="firewall-rules-list" />
</template>

<script setup lang="ts">
import PageHeader from "@/components/PageHeader.vue";
import FirewallRulesList from "../components/FirewallRules/FirewallRulesList.vue";
</script>
