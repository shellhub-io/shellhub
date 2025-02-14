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

<script lang="ts">
import { computed, defineComponent, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { useStore } from "../../store";
import DataTable from "../DataTable.vue";
import showTag from "../../hooks/tag";
import displayOnlyTenCharacters from "../../hooks/string";
import { filterType } from "../../interfaces/IFirewallRule";
import { INotificationsError } from "../../interfaces/INotifications";

export default defineComponent({
  setup() {
    const store = useStore();
    const router = useRouter();
    const loading = ref(false);
    const page = ref(1);
    const itemsPerPage = ref(10);

    onMounted(() => {
      try {
        loading.value = true;
        store.dispatch("firewallRules/fetch", {
          page: page.value,
          perPage: itemsPerPage.value,
        });
      } catch {
        store.dispatch("snackbar/showSnackbarErrorAction", INotificationsError.firewallRuleList);
      } finally {
        loading.value = false;
      }
    });

    const numberFirewalls = computed(() => store.getters["firewallRules/numberFirewalls"]);

    const getFirewallRules = async (perPagaeValue: number, pageValue: number) => {
      try {
        loading.value = true;
        const hasFirewallRules = await store.dispatch("firewallRules/fetch", {
          page: pageValue,
          perPage: perPagaeValue,
        });
        if (!hasFirewallRules) page.value--;
      } catch {
        store.dispatch("snackbar/showSnackbarErrorAction", INotificationsError.firewallRuleList);
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
        store.dispatch("snackbar/setSnackbarErrorDefault");
      }
    };

    const changeItemsPerPage = async (newItemsPerPage: number) => {
      itemsPerPage.value = newItemsPerPage;
    };

    watch(itemsPerPage, () => {
      getFirewallRules(itemsPerPage.value, page.value);
    });

    const firewallRules = computed(() => store.getters["firewallRules/list"]);

    const formatSourceIP = (ip: string) => (ip === ".*" ? "Any IP" : ip);

    const formatUsername = (username: string) => username === ".*" ? "All users" : username;

    const formatHostnameFilter = (filter: filterType) => filter.hostname === ".*" ? "All devices" : filter.hostname;

    const isHostname = (filter: filterType) => Object.prototype.hasOwnProperty.call(filter, "hostname");

    const goToFirewallRule = (ruleId : string) => router.push({ name: "firewallRulesDetails", params: { id: ruleId } });

    return {
      headers: [
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
      ],
      loading,
      page,
      itemsPerPage,
      firewallRules,
      numberFirewalls,
      next,
      prev,
      changeItemsPerPage,
      formatSourceIP,
      formatUsername,
      formatHostnameFilter,
      isHostname,
      showTag,
      displayOnlyTenCharacters,
      goToFirewallRule,
    };
  },
  components: { DataTable },
});
</script>
