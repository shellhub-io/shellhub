<template>
  <div data-test="firewallRules-list">
    <DataTable
      :headers="headers"
      :items="firewallRules"
      :itemsPerPage="itemsPerPage"
      :loading="loading"
      :page="page"
      :actualPage="page"
      :totalCount="numberFirewalls"
      @changeItemsPerPage="changeItemsPerPage"
      @clickNextPage="next"
      @clickPreviousPage="prev"
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
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import useSnackbarStore from "@admin/store/modules/snackbar";
import useFirewallRulesStore from "@admin/store/modules/firewall_rules";
import DataTable from "../DataTable.vue";
import showTag from "../../hooks/tag";
import displayOnlyTenCharacters from "../../hooks/string";
import { filterType } from "../../interfaces/IFirewallRule";
import { INotificationsError } from "../../interfaces/INotifications";

const router = useRouter();
const snackbarStore = useSnackbarStore();
const firewallRulesStore = useFirewallRulesStore();
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
onMounted(() => {
  try {
    loading.value = true;
    firewallRulesStore.fetch({
      page: page.value,
      perPage: itemsPerPage.value,
    });
  } catch {
    snackbarStore.showSnackbarErrorAction(INotificationsError.firewallRuleList);
  } finally {
    loading.value = false;
  }
});

const numberFirewalls = computed(() => firewallRulesStore.getNumberFirewalls);

const getFirewallRules = async (perPageValue: number, pageValue: number) => {
  try {
    loading.value = true;
    const hasFirewallRules = await firewallRulesStore.fetch({
      page: pageValue,
      perPage: perPageValue,
    });
    if (!hasFirewallRules) page.value--;
  } catch {
    snackbarStore.showSnackbarErrorAction(INotificationsError.firewallRuleList);
  } finally {
    loading.value = false;
  }
};

const next = async () => {
  await getFirewallRules(itemsPerPage.value, ++page.value);
};

const prev = async () => {
  try {
    if (page.value > 1) await getFirewallRules(itemsPerPage.value, --page.value);
  } catch (error) {
    snackbarStore.showSnackbarErrorDefault();
  }
};

const changeItemsPerPage = async (newItemsPerPage: number) => {
  itemsPerPage.value = newItemsPerPage;
};

watch(itemsPerPage, () => {
  getFirewallRules(itemsPerPage.value, page.value);
});

const firewallRules = computed(() => firewallRulesStore.list);

const formatSourceIP = (ip: string) => (ip === ".*" ? "Any IP" : ip);

const formatUsername = (username: string) => username === ".*" ? "All users" : username;

const formatHostnameFilter = (filter: filterType) => filter.hostname === ".*" ? "All devices" : filter.hostname;

const isHostname = (filter: filterType) => Object.prototype.hasOwnProperty.call(filter, "hostname");

const goToFirewallRule = (ruleId : string) => router.push({ name: "firewallRulesDetails", params: { id: ruleId } });

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
