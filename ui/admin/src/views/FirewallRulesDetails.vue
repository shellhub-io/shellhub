<template>
  <h1>Firewall Details</h1>
  <v-card class="mt-2 pa-4 bg-background border">
    <v-card-text v-if="!isFirewallRuleEmpty">
      <div>
        <h3 class="text-overline">id:</h3>
        <p :data-test="firewallRule.id">{{ firewallRule.id }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Tenant Id:</h3>
        <p :data-test="firewallRule.tenant_id">{{ firewallRule.tenant_id }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Priority:</h3>
        <p :data-test="firewallRule.priority">{{ firewallRule.priority }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Action:</h3>
        <p :data-test="firewallRule.action" class="text-capitalize">{{ firewallRule.action }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Source Ip:</h3>
        <p :data-test="firewallRule.source_ip">{{ formatSourceIP(firewallRule.source_ip) }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Username:</h3>
        <p :data-test="firewallRule.username">{{ formatUsername(firewallRule.username) }}</p>
      </div>

      <div v-if="firewallRule.filter">
        <h3 class="text-overline mt-3">Filter:</h3>
        <p v-if="isHostname(firewallRule.filter)" data-test="firewall-rule-filter">
          {{ formatHostnameFilter(firewallRule.filter) }}
        </p>
        <div v-else :data-test="firewallRule.filter">
          <v-tooltip
            v-for="(tag, index) in firewallRule.filter.tags"
            :key="index"
            bottom
            :disabled="!showTag(tag.name)"
          >
            <template #activator="{ props }">
              <v-chip class="mr-1" density="compact" outlined v-bind="props">
                {{ displayOnlyTenCharacters(tag.name) }}
              </v-chip>
            </template>

            <span>
              {{ tag.name }}
            </span>
          </v-tooltip>
        </div>
      </div>
    </v-card-text>
    <p v-else class="text-center">Something is wrong, try again!</p>
  </v-card>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import useFirewallRulesStore from "@admin/store/modules/firewall_rules";
import isHostname from "@/utils/isHostname";
import useSnackbar from "@/helpers/snackbar";
import { IAdminFirewallRule } from "../interfaces/IFirewallRule";
import showTag from "@/utils/tag";
import { displayOnlyTenCharacters, formatHostnameFilter, formatSourceIP, formatUsername } from "@/utils/string";

const route = useRoute();
const snackbar = useSnackbar();
const firewallRulesStore = useFirewallRulesStore();

const firewallRuleId = computed(() => route.params.id as string);
const firewallRule = ref({} as IAdminFirewallRule);

onMounted(async () => {
  try {
    firewallRule.value = await firewallRulesStore.fetchFirewallRuleById(firewallRuleId.value);
  } catch {
    snackbar.showError("Failed to get firewall rule details.");
  }
});

const isFirewallRuleEmpty = computed(() => !firewallRule.value || !firewallRule.value.id);

defineExpose({ firewallRule });
</script>
