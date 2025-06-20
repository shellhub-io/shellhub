<template>
  <DataTable
    :headers
    :items="firewallRules"
    v-model:itemsPerPage="itemsPerPage"
    v-model:page="page"
    :loading
    :totalCount="firewallRulesCount"
    :itemsPerPageOptions="[10, 20, 50, 100]"
    data-test="firewall-rules-list"
  >
    <template v-slot:rows>
      <tr v-for="(firewallRule, index) in firewallRules" :key="index">
        <td>
          {{ firewallRule.tenant_id }}
        </td>
        <td>
          {{ firewallRule.priority }}
        </td>
        <td>
          {{ firewallRule.action }}
        </td>
        <td>
          {{ formatSourceIP(firewallRule.source_ip) }}
        </td>
        <td>
          {{ formatUsername(firewallRule.username) }}
        </td>
        <td>
          <div v-if="isHostname(firewallRule.filter)">
            {{ formatHostnameFilter(firewallRule.filter) }}
          </div>
          <div v-else>
            <v-tooltip
              v-for="(tag, index) in firewallRule.filter.tags"
              :key="index"
              bottom
              :disabled="!showTag(tag)"
            >
              <template #activator="{ props }">
                <v-chip
                  class="mr-1"
                  density="compact"
                  outlined
                  v-bind="props"
                >
                  {{ displayOnlyTenCharacters(tag) }}
                </v-chip>
              </template>

              <span v-if="showTag(tag)">
                {{ tag }}
              </span>
            </v-tooltip>
          </div>
        </td>
        <td>
          <v-tooltip bottom anchor="bottom">
            <template v-slot:activator="{ props }">
              <v-icon
                tag="a"
                dark
                v-bind="props"
                @click="goToFirewallRule(firewallRule.id)"
                @keypress.enter="goToFirewallRule(firewallRule.id)"
                tabindex="0"
              >
                mdi-information
              </v-icon>
            </template>
            <span>Details</span>
          </v-tooltip>
        </td>
      </tr>
    </template>
  </DataTable>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import useFirewallRulesStore from "@admin/store/modules/firewall_rules";
import useSnackbar from "@/helpers/snackbar";
import DataTable from "@/components/DataTable.vue";
import showTag from "../../hooks/tag";
import displayOnlyTenCharacters from "../../hooks/string";
import { filterType } from "../../interfaces/IFirewallRule";

const router = useRouter();
const snackbar = useSnackbar();
const firewallRulesStore = useFirewallRulesStore();
const firewallRules = computed(() => firewallRulesStore.list);
const firewallRulesCount = computed(() => firewallRulesStore.getNumberFirewalls);
const loading = ref(false);
const page = ref(1);
const itemsPerPage = ref(10);
const headers = ref([
  {
    text: "Tenant Id",
    value: "tenant_id",
  },
  {
    text: "Priority",
    value: "priority",
  },
  {
    text: "Action",
    value: "action",
  },
  {
    text: "Source Ip",
    value: "source_ip",
  },
  {
    text: "Username",
    value: "username",
  },
  {
    text: "Filter",
    value: "filter",
  },
  {
    text: "Actions",
    value: "actions",
  },
]);

const formatSourceIP = (ip: string) => (ip === ".*" ? "Any IP" : ip);

const formatUsername = (username: string) => username === ".*" ? "All users" : username;

const formatHostnameFilter = (filter: filterType) => filter.hostname === ".*" ? "All devices" : filter.hostname;

const isHostname = (filter: filterType) => Object.prototype.hasOwnProperty.call(filter, "hostname");

const goToFirewallRule = (ruleId: string) => router.push({ name: "firewallRulesDetails", params: { id: ruleId } });

const fetchFirewallRules = async () => {
  try {
    loading.value = true;
    await firewallRulesStore.fetch({
      page: page.value,
      perPage: itemsPerPage.value,
    });
  } catch {
    snackbar.showError("Failed to fetch firewall rules.");
  }
  loading.value = false;
};

watch([itemsPerPage, page], () => {
  fetchFirewallRules();
});

onMounted(async () => {
  await fetchFirewallRules();
});

defineExpose({
  headers,
  loading,
  itemsPerPage,
  page,
  firewallRules,
  formatSourceIP,
  formatUsername,
  formatHostnameFilter,
  displayOnlyTenCharacters,
  showTag,
});
</script>
