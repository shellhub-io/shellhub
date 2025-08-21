<template>
  <div>
    <DataTable
      v-model:page="page"
      v-model:itemsPerPage="itemsPerPage"
      :headers
      :items="publicKeys"
      :totalCount="publicKeyCount"
      :loading
      :itemsPerPageOptions="[10, 20, 50, 100]"
      data-test="public-keys-list"
    >
      <template v-slot:rows>
        <tr v-for="(item, i) in publicKeys" :key="i" data-test="public-key-item">
          <td class="text-center" data-test="public-key-name">
            {{ item.name }}
          </td>

          <td class="text-center" data-test="public-key-fingerprint">
            {{ item.fingerprint }}
          </td>

          <td class="text-center" data-test="public-key-filter">
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

          <td class="text-center" data-test="public-key-username">
            {{ formatUsername(item.username) }}
          </td>

          <td class="text-center" data-test="public-key-created-at">
            {{ formatAbbreviatedDateTime(item.created_at) }}
          </td>

          <td class="text-center" data-test="public-key-actions">
            <v-menu location="bottom" scrim eager>
              <template v-slot:activator="{ props }">
                <v-btn
                  v-bind="props"
                  variant="plain"
                  class="border rounded bg-v-theme-background"
                  density="comfortable"
                  size="default"
                  icon="mdi-format-list-bulleted"
                  data-test="public-key-actions"
                />
              </template>
              <v-list class="bg-v-theme-surface" lines="two" density="compact">
                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="hasAuthorizationFormDialogEdit"
                >
                  <template v-slot:activator="{ props }">
                    <div v-bind="props">
                      <PublicKeyEdit
                        :publicKey="item"
                        :hasAuthorization="hasAuthorizationFormDialogEdit"
                        @update="refreshPublicKeys"
                      />
                    </div>
                  </template>
                  <span> You don't have this kind of authorization. </span>
                </v-tooltip>

                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="hasAuthorizationFormDialogRemove"
                >
                  <template v-slot:activator="{ props }">
                    <div v-bind="props">
                      <PublicKeyDelete
                        :fingerprint="item.fingerprint"
                        :hasAuthorization="hasAuthorizationFormDialogRemove"
                        @update="refreshPublicKeys"
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
import { actions, authorizer } from "@/authorizer";
import { Filter, HostnameFilter } from "@/interfaces/IFilter";
import hasPermission from "@/utils/permission";
import {
  displayOnlyTenCharacters,
  formatHostnameFilter,
  formatUsername,
} from "@/utils/string";
import { formatAbbreviatedDateTime } from "@/utils/date";
import showTag from "@/utils/tag";
import DataTable from "../DataTable.vue";
import PublicKeyDelete from "./PublicKeyDelete.vue";
import PublicKeyEdit from "./PublicKeyEdit.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useAuthStore from "@/store/modules/auth";
import usePublicKeysStore from "@/store/modules/public_keys";

const headers = [
  {
    text: "Name",
    value: "name",
  },
  {
    text: "Fingerprint",
    value: "fingerprint",
  },
  {
    text: "Filter",
    value: "filter",
  },
  {
    text: "Username",
    value: "username",
  },
  {
    text: "Created At",
    value: "created_at",
  },
  {
    text: "Actions",
    value: "actions",
  },
];
const authStore = useAuthStore();
const publicKeysStore = usePublicKeysStore();
const snackbar = useSnackbar();
const loading = ref(false);
const itemsPerPage = ref(10);
const page = ref(1);
const publicKeys = computed(() => publicKeysStore.publicKeys);
const publicKeyCount = computed(() => publicKeysStore.publicKeyCount);
const hasAuthorizationFormDialogEdit = computed(() => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.publicKey.edit);
});

const hasAuthorizationFormDialogRemove = computed(() => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.publicKey.remove);
});

const getPublicKeysList = async () => {
  try {
    loading.value = true;
    await publicKeysStore.fetchPublicKeyList({
      page: page.value,
      perPage: itemsPerPage.value,
    });
    loading.value = false;
  } catch (error: unknown) {
    snackbar.showError("Failed to load public keys.");
    handleError(error);
  }
};

watch([page, itemsPerPage], async () => {
  await getPublicKeysList();
});

const refreshPublicKeys = async () => {
  await getPublicKeysList();
};

const isHostname = (filter: Filter): filter is HostnameFilter => "hostname" in filter;

defineExpose({ publicKeys, hasAuthorizationFormDialogEdit, hasAuthorizationFormDialogRemove });
</script>
