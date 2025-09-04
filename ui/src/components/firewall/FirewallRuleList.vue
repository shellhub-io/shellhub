<template>
  <div>
    <DataTable
      v-model:page="page"
      v-model:itemsPerPage="itemsPerPage"
      :headers
      :items="firewallRules"
      :totalCount="firewallRuleCount"
      :loading
      :itemsPerPageOptions="[10, 20, 50, 100]"
      data-test="firewallRules-list"
    >
      <template v-slot:rows>
        <tr v-for="(item, i) in firewallRules" :key="i">
          <td class="text-center">
            <v-icon
              data-test="firewall-rules-active"
              :color="item.active ? 'success' : ''"
              icon="mdi-check-circle" />
          </td>

          <td class="text-center" data-test="firewall-rules-priority">{{ item.priority }}</td>

          <td class="text-center" data-test="firewall-rules-action">
            {{ capitalizeText(item.action) }}
          </td>

          <td class="text-center" data-test="firewall-rules-source-ip">
            {{ formatSourceIP(item.source_ip) }}
          </td>

          <td class="text-center" data-test="firewall-rules-username">
            {{ formatUsername(item.username) }}
          </td>

          <td class="text-center" data-test="firewall-rules-filter">
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
                <v-btn
                  v-bind="props"
                  variant="plain"
                  class="border rounded bg-v-theme-background"
                  density="comfortable"
                  size="default"
                  icon="mdi-format-list-bulleted"
                  data-test="firewall-rules-actions"
                />
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
                        :hasAuthorization="hasAuthorizationFormDialogEdit()"
                        @update="refreshFirewallRules"
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
                        v-if="item.id"
                        :id="item.id"
                        @update="refreshFirewallRules"
                        :hasAuthorization="hasAuthorizationFormDialogEdit()"
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

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import axios, { AxiosError } from "axios";
import { actions, authorizer } from "@/authorizer";
import isHostname from "@/utils/isHostname";
import { capitalizeText, displayOnlyTenCharacters, formatHostnameFilter, formatSourceIP, formatUsername } from "@/utils/string";
import showTag from "@/utils/tag";
import hasPermission from "@/utils/permission";
import DataTable from "../DataTable.vue";
import FirewallRuleDelete from "./FirewallRuleDelete.vue";
import FirewallRuleEdit from "./FirewallRuleEdit.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useAuthStore from "@/store/modules/auth";
import useFirewallRulesStore from "@/store/modules/firewall_rules";

const headers = [
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
];

const authStore = useAuthStore();
const firewallRulesStore = useFirewallRulesStore();
const snackbar = useSnackbar();
const loading = ref(false);
const itemsPerPage = ref(10);
const page = ref(1);
const firewallRules = computed(() => firewallRulesStore.firewallRules);
const firewallRuleCount = computed(() => firewallRulesStore.firewallRuleCount);

const getFirewalls = async () => {
  try {
    loading.value = true;
    await firewallRulesStore.fetchFirewallRuleList({
      perPage: itemsPerPage.value,
      page: page.value,
    });
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response?.status === 403) {
        snackbar.showError("You don't have permission to access this resource.");
        handleError(error);
      }
    } else {
      snackbar.showError("An error occurred while loading the firewall rules.");
      handleError(error);
    }
  } finally {
    loading.value = false;
  }
};

watch([page, itemsPerPage], async () => {
  await getFirewalls();
});

const refreshFirewallRules = async () => {
  try {
    await getFirewalls();
  } catch (error: unknown) {
    snackbar.showError("An error occurred while refreshing the firewall rules.");
    handleError(error);
  }
};

const hasAuthorizationFormDialogEdit = () => hasPermission(authorizer.role[authStore.role], actions.firewall.edit);

const hasAuthorizationFormDialogRemove = () => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.firewall.remove);
};
</script>
