<template>
  <div>
    <DataTable
      :headers="headers"
      :items="firewallRules"
      :itemsPerPage="itemsPerPage"
      :nextPage="next"
      :previousPage="prev"
      :loading="loading"
      :totalCount="getNumberFirewallRules"
      :actualPage="page"
      :comboboxOptions="[10, 20, 50, 100]"
      @changeItemsPerPage="changeItemsPerPage"
      @clickNextPage="next"
      @clickPreviousPage="prev"
      data-test="firewallRules-list"
    >
      <template v-slot:rows>
        <tr v-for="(item, i) in firewallRules" :key="i">
          <td class="text-center">
            <v-icon v-if="item.active" color="success">
              mdi-check-circle
            </v-icon>
            <v-tooltip location="bottom" v-else>
              <template v-slot:activator="{ props }">
                <v-icon v-bind="props"> mdi-check-circle </v-icon>
              </template>
              <span>{{ lastSeen(item.last_seen) }}</span>
            </v-tooltip>
          </td>

          <td class="text-center">{{ item.priority }}</td>

          <td class="text-center">
            {{ capitalizeText(item.action) }}
          </td>

          <td class="text-center">
            {{ formatSourceIP(item.source_ip) }}
          </td>

          <td class="text-center">
            {{ formatUsername(item.username) }}
          </td>

          <td class="text-center">
            <div v-if="isHostname(item.filter)">
              {{ formatHostnameFilter(item.filter) }}
            </div>
            <div v-else>
              <v-tooltip
                v-for="(tag, index) in item.filter.tags"
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

          <td class="text-center">
            <v-menu location="bottom" scrim eager>
              <template v-slot:activator="{ props }">
                <v-chip v-bind="props" density="comfortable" size="small">
                  <v-icon>mdi-dots-horizontal</v-icon>
                </v-chip>
              </template>
              <v-list class="bg-v-theme-surface" lines="two" density="compact">
                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="hasAuthorizationFormDialogEdit()"
                >
                  <template v-slot:activator="{ props }">
                    <div v-bind="props">
                      <FirewallRuleEdit
                        :firewallRule="item"
                        :notHasAuthorization="!hasAuthorizationFormDialogEdit()"
                        @update="refreshFirewallRules"
                        data-test="firewallRuleEdit-component"
                      />
                    </div>
                  </template>
                  <span> You don't have this kind of authorization. </span>
                </v-tooltip>

                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="hasAuthorizationFormDialogRemove()"
                >
                  <template v-slot:activator="{ props }">
                    <div v-bind="props">
                      <FirewallRuleDelete
                        :id="item.id"
                        @update="refreshFirewallRules"
                        :notHasAuthorization="!hasAuthorizationFormDialogEdit()"
                        data-test="firewallRuleRemove-component"
                      />
                    </div>
                  </template>
                  <span> You don't have this kind of authorization. </span>
                </v-tooltip>
              </v-list>
            </v-menu>
          </td>
        </tr>
      </template>
    </DataTable>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, ref, watch } from "vue";
import { actions, authorizer } from "../../authorizer";
import { filterType } from "../../interfaces/IFirewallRule";
import { useStore } from "../../store";
import { lastSeen } from "../../utils/formateDate";
import { capitalizeText, displayOnlyTenCharacters } from "../../utils/string";
import showTag from "../../utils/tag";
import hasPermission from "../../utils/permission";
import DataTable from "../DataTable.vue";
import FirewallRuleDelete from "./FirewallRuleDelete.vue";
import FirewallRuleEdit from "./FirewallRuleEdit.vue";
import { INotificationsError } from "../../interfaces/INotifications";

export default defineComponent({
  setup() {
    const store = useStore();
    const loading = ref(false);
    const itemsPerPage = ref(10);
    const page = ref(1);

    const firewallRules = computed(() => store.getters["firewallRules/list"]);

    const getNumberFirewallRules = computed(
      () => store.getters["firewallRules/getNumberFirewalls"],
    );

    const getFirewalls = async (perPagaeValue: number, pageValue: number) => {
      if (!store.getters["boxs/getStatus"]) {
        const data = {
          perPage: perPagaeValue,
          page: pageValue,
        };

        try {
          loading.value = true;
          const hasRules = await store.dispatch("firewallRules/fetch", data);
          if (!hasRules) {
            page.value--;
          }
        } catch (error: any) {
          if (error.response.status === 403) {
            store.dispatch("snackbar/showSnackbarErrorAssociation");
            throw new Error(error);
          } else {
            store.dispatch(
              "snackbar/showSnackbarErrorLoading",
              INotificationsError.firewallRuleList,
            );
            throw new Error(error);
          }
        } finally {
          loading.value = false;
        }
      } else {
        // setArrays();
        store.dispatch("boxs/setStatus", false);
      }
    };

    onMounted(() => {
      getFirewalls(itemsPerPage.value, page.value);
    });

    const next = async () => {
      await getFirewalls(itemsPerPage.value, ++page.value);
    };

    const prev = async () => {
      try {
        if (page.value > 1) await getFirewalls(itemsPerPage.value, --page.value);
      } catch (error) {
        store.dispatch("snackbar/setSnackbarErrorDefault");
      }
    };

    const changeItemsPerPage = async (newItemsPerPage: number) => {
      itemsPerPage.value = newItemsPerPage;
    };

    watch(itemsPerPage, async () => {
      await getFirewalls(itemsPerPage.value, page.value);
    });

    const refreshFirewallRules = async () => {
      try {
        await store.dispatch("firewallRules/refresh");
        getFirewalls(itemsPerPage.value, page.value);
      } catch (error: any) {
        store.dispatch("snackbar/setSnackbarErrorDefault");
        throw new Error(error);
      }
    };

    const formatSourceIP = (ip: string) => (ip === ".*" ? "Any IP" : ip);

    const formatUsername = (username: string) => username === ".*" ? "All users" : username;

    const formatHostnameFilter = (filter: filterType) => filter.hostname === ".*" ? "All devices" : filter.hostname;

    const isHostname = (filter: filterType) => Object.prototype.hasOwnProperty.call(filter, "hostname");

    const hasAuthorizationFormDialogEdit = () => {
      const role = store.getters["auth/role"];
      if (role !== "") {
        return hasPermission(authorizer.role[role], actions.firewall.edit);
      }

      return false;
    };

    const hasAuthorizationFormDialogRemove = () => {
      const role = store.getters["auth/role"];
      if (role !== "") {
        return hasPermission(authorizer.role[role], actions.firewall.remove);
      }

      return false;
    };

    return {
      headers: [
        {
          text: "Active",
          value: "active",
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
          text: "Source IP",
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
      firewallRules,
      getNumberFirewallRules,
      loading,
      itemsPerPage,
      page,
      next,
      prev,
      changeItemsPerPage,
      refreshFirewallRules,
      formatSourceIP,
      formatUsername,
      formatHostnameFilter,
      isHostname,
      lastSeen,
      capitalizeText,
      displayOnlyTenCharacters,
      showTag,
      hasAuthorizationFormDialogEdit,
      hasAuthorizationFormDialogRemove,
    };
  },
  components: {
    DataTable,
    FirewallRuleDelete,
    FirewallRuleEdit,
  },
});
</script>
