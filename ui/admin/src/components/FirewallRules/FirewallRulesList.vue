<template>
  <DataTable
    v-model:items-per-page="itemsPerPage"
    v-model:page="page"
    :headers
    :items="firewallRules"
    :loading
    :total-count="firewallRulesCount"
    :items-per-page-options="[10, 20, 50, 100]"
    data-test="firewall-rules-list"
  >
    <template #rows>
      <tr
        v-for="firewallRule in firewallRules"
        :key="firewallRule.id"
      >
        <td>
          {{ firewallRule.tenant_id }}
        </td>
        <td>
          {{ firewallRule.priority }}
        </td>
        <td class="text-capitalize">
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
              :disabled="!showTag(tag.name)"
            >
              <template #activator="{ props }">
                <v-chip
                  class="mr-1"
                  density="compact"
                  outlined
                  v-bind="props"
                >
                  {{ displayOnlyTenCharacters(tag.name) }}
                </v-chip>
              </template>

              <span>
                {{ tag.name }}
              </span>
            </v-tooltip>
          </div>
        </td>
        <td>
          <v-tooltip
            bottom
            anchor="bottom"
          >
            <template #activator="{ props }">
              <v-icon
                tag="a"
                dark
                v-bind="props"
                tabindex="0"
                icon="mdi-information"
                @click="goToFirewallRule(firewallRule.id)"
                @keypress.enter="goToFirewallRule(firewallRule.id)"
              />
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
import isHostname from "@/utils/isHostname";
import useSnackbar from "@/helpers/snackbar";
import DataTable from "@/components/Tables/DataTable.vue";
import showTag from "@/utils/tag";
import { displayOnlyTenCharacters, formatHostnameFilter, formatSourceIP, formatUsername } from "@/utils/string";
import handleError from "@/utils/handleError";

const router = useRouter();
const snackbar = useSnackbar();
const firewallRulesStore = useFirewallRulesStore();
const firewallRules = computed(() => firewallRulesStore.firewallRules);
const firewallRulesCount = computed(() => firewallRulesStore.firewallRulesCount);
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

const goToFirewallRule = (ruleId: string) => router.push({ name: "firewallRulesDetails", params: { id: ruleId } });

const fetchFirewallRules = async () => {
  try {
    loading.value = true;
    await firewallRulesStore.fetchFirewallRulesList({
      page: page.value,
      perPage: itemsPerPage.value,
    });
  } catch (error) {
    handleError(error);
    snackbar.showError("Failed to fetch firewall rules.");
  }
  loading.value = false;
};

watch([itemsPerPage, page], async () => {
  await fetchFirewallRules();
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
