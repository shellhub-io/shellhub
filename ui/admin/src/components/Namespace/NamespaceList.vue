<template>
  <div>
    <DataTable
      :headers
      :items="namespaces"
      v-model:itemsPerPage="itemsPerPage"
      v-model:page="page"
      :loading
      :totalCount="numberOfNamespaces"
      :itemsPerPageOptions="[10, 20, 50, 100]"
      data-test="namespaces-list"
    >
      <template v-slot:rows>
        <tr v-for="(namespace, i) in namespaces" :key="i">
          <td>
            {{ namespace.name }}
          </td>
          <td>
            {{ sumDevicesCount(namespace) }}
          </td>
          <td>
            {{ namespace.tenant_id }}
          </td>
          <td>
            {{ namespace.owner }}
          </td>
          <td>
            <div v-if="namespace.settings">
              {{ namespace.settings.session_record }}
            </div>
          </td>
          <td>
            <v-tooltip bottom anchor="bottom">
              <template v-slot:activator="{ props }">
                <v-icon
                  tag="a"
                  dark
                  v-bind="props"
                  @click="goToNamespace(namespace.tenant_id)"
                  @keypress.enter="goToNamespace(namespace.tenant_id)"
                  tabindex="0"
                >mdi-information
                </v-icon>
              </template>
              <span>Details</span>
            </v-tooltip>

            <NamespaceEdit :namespace="namespace" />
          </td>
        </tr>
      </template>
    </DataTable>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import useNamespacesStore from "@admin/store/modules/namespaces";
import { INamespace } from "@admin/interfaces/INamespace";
import useSnackbar from "@/helpers/snackbar";
import DataTable from "@/components/DataTable.vue";
import NamespaceEdit from "./NamespaceEdit.vue";

const snackbar = useSnackbar();
const namespacesStore = useNamespacesStore();
const router = useRouter();
const loading = ref(false);
const page = ref(1);
const itemsPerPage = ref(10);
const filter = ref("");

const headers = ref([
  {
    text: "Name",
    value: "name",
  },
  {
    text: "Devices",
    value: "devices",
  },
  {
    text: "Tenant ID",
    value: "tenant_id",
  },
  {
    text: "Owner",
    value: "owner",
  },
  {
    text: "Session Record",
    value: "settings",
  },
  {
    text: "Actions",
    value: "actions",
  },
]);

onMounted(async () => {
  try {
    loading.value = true;
    await namespacesStore.fetch({
      perPage: itemsPerPage.value,
      page: page.value,
      filter: filter.value,
    });
  } catch {
    snackbar.showError("Failed to fetch namespaces.");
  } finally {
    loading.value = false;
  }
});

const namespaces = computed(() => namespacesStore.list);
const numberOfNamespaces = computed(() => namespacesStore.getnumberOfNamespaces);

const goToNamespace = (namespace: string) => {
  router.push({ name: "namespaceDetails", params: { id: namespace } });
};

const getNamespaces = async (perPageValue: number, pageValue: number) => {
  try {
    loading.value = true;
    await namespacesStore.fetch({
      page: pageValue,
      perPage: perPageValue,
      filter: filter.value,
    });
  } catch {
    snackbar.showError("Failed to fetch namespaces.");
  } finally {
    loading.value = false;
  }
};

watch([itemsPerPage, page], async () => {
  await getNamespaces(itemsPerPage.value, page.value);
});

const sumDevicesCount = (namespace: INamespace) => {
  const { devices_accepted_count: acceptedCount, devices_pending_count: pendingCount, devices_rejected_count: rejectedCount } = namespace;
  return (acceptedCount + pendingCount + rejectedCount) || 0;
};
</script>
