<template>
  <div class="d-flex pa-0 align-center"><h1>Firewall Rule Details</h1></div>
  <v-card
    v-if="firewallRule.id"
    class="mt-2 border rounded bg-background"
  >
    <v-card-title class="pa-4 d-flex align-center justify-space-between bg-v-theme-surface">
      <div class="d-flex align-center ml-2 ga-3">
        <v-tooltip
          location="bottom"
          :text="firewallRule.active ? 'Active' : 'Inactive'"
        >
          <template #activator="{ props }">
            <v-icon
              v-bind="props"
              :color="firewallRule.active ? 'success' : '#E53935'"
              data-test="active-icon"
              :icon="firewallRule.active ? 'mdi-shield-check' : 'mdi-shield-off'"
            />
          </template>
        </v-tooltip>
        <h2 class="text-h6">Rule #{{ firewallRule.priority }}</h2>
        <v-chip
          size="small"
          data-test="firewall-action-chip"
          class="text-capitalize"
          :color="firewallRule.action === 'allow' ? 'success' : 'error'"
          :text="firewallRule.action"
        />
      </div>
    </v-card-title>

    <v-divider />

    <v-card-text class="pa-4 pt-0">
      <v-row class="py-3">
        <v-col
          cols="12"
          md="6"
          class="my-0 py-0"
        >
          <div data-test="firewall-id-field">
            <h3 class="item-title">ID:</h3>
            <p class="text-truncate">{{ firewallRule.id }}</p>
          </div>

          <div data-test="firewall-tenant-field">
            <h3 class="item-title">Namespace:</h3>
            <router-link
              :to="{ name: 'namespaceDetails', params: { id: firewallRule.tenant_id } }"
              class="hyper-link"
            >
              {{ firewallRule.tenant_id }}
            </router-link>
          </div>

          <div data-test="firewall-priority-field">
            <h3 class="item-title">Priority:</h3>
            <p>{{ firewallRule.priority }}</p>
          </div>
        </v-col>

        <v-col
          cols="12"
          md="6"
          class="my-0 py-0"
        >
          <div data-test="firewall-source-ip-field">
            <h3 class="item-title">Source IP:</h3>
            <code>{{ formatSourceIP(firewallRule.source_ip) }}</code>
          </div>

          <div data-test="firewall-username-field">
            <h3 class="item-title">Username:</h3>
            <p>{{ formatUsername(firewallRule.username) }}</p>
          </div>

          <div
            v-if="firewallRule.filter"
            data-test="firewall-filter-field"
          >
            <h3 class="item-title">Filter:</h3>
            <p v-if="isHostname(firewallRule.filter)">{{ formatHostnameFilter(firewallRule.filter) }}</p>
            <div v-else>
              <v-tooltip
                v-for="(tag, index) in firewallRule.filter.tags"
                :key="index"
                bottom
                :disabled="!showTag(tag.name)"
                :text="tag.name"
              >
                <template #activator="{ props }">
                  <v-chip
                    size="small"
                    v-bind="props"
                    class="mr-2"
                  >
                    {{ displayOnlyTenCharacters(tag.name) }}
                  </v-chip>
                </template>
              </v-tooltip>
            </div>
          </div>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
  <v-card
    v-else
    class="mt-2 pa-4 bg-v-theme-surface"
  >
    <p class="text-center">Something is wrong, try again!</p>
  </v-card>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useRoute } from "vue-router";
import useFirewallRulesStore from "@admin/store/modules/firewall_rules";
import isHostname from "@/utils/isHostname";
import useSnackbar from "@/helpers/snackbar";
import showTag from "@/utils/tag";
import { displayOnlyTenCharacters, formatHostnameFilter, formatSourceIP, formatUsername } from "@/utils/string";

const route = useRoute();
const snackbar = useSnackbar();
const firewallRulesStore = useFirewallRulesStore();

const firewallRuleId = computed(() => route.params.id as string);
const firewallRule = computed(() => firewallRulesStore.firewallRule);

onMounted(async () => {
  try {
    await firewallRulesStore.fetchFirewallRuleById(firewallRuleId.value);
  } catch {
    snackbar.showError("Failed to get firewall rule details.");
  }
});
</script>

<style scoped>
.hyper-link {
  color: inherit;
  text-decoration: underline;
}

.hyper-link:visited,
.hyper-link:hover,
.hyper-link:active {
  color: inherit;
}
</style>
