<template>
  <div>
    <DataTable
      v-model:page="page"
      v-model:itemsPerPage="itemsPerPage"
      :headers
      :items="publicKeys"
      :totalCount="publicKeysCount"
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
import { useStore } from "@/store";
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
const store = useStore();
const authStore = useAuthStore();
const snackbar = useSnackbar();
const loading = ref(false);
const itemsPerPage = ref(10);
const page = ref(1);
const publicKeys = computed(() => store.getters["publicKeys/list"]);
const publicKeysCount = computed(
  () => store.getters["publicKeys/getNumberPublicKeys"],
);
const hasAuthorizationFormDialogEdit = computed(() => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.publicKey.edit);
});

const hasAuthorizationFormDialogRemove = computed(() => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.publicKey.remove);
});

const getPublicKeysList = async (
  perPageValue: number,
  pageValue: number,
) => {
  const data = {
    perPage: perPageValue,
    page: pageValue,
  };
  try {
    loading.value = true;
    const hasPublicKeys = await store.dispatch("publicKeys/fetch", data);

    if (!hasPublicKeys) {
      page.value--;
    }
    loading.value = false;
  } catch (error: unknown) {
    snackbar.showError("Failed to load public keys.");
    handleError(error);
  }
};

watch([page, itemsPerPage], async () => {
  await getPublicKeysList(itemsPerPage.value, page.value);
});

const refreshPublicKeys = async () => {
  await store.dispatch("publicKeys/refresh");
};

const isHostname = (filter: Filter): filter is HostnameFilter => "hostname" in filter;

defineExpose({ publicKeys, hasAuthorizationFormDialogEdit, hasAuthorizationFormDialogRemove });
</script>
