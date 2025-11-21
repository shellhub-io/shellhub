<template>
  <div>
    <DataTable
      v-model:items-per-page="itemsPerPage"
      v-model:page="page"
      :headers="headers"
      :items="namespaces"
      :loading="loading"
      :total-count="namespaceCount"
      :items-per-page-options="[10, 20, 50, 100]"
      data-test="namespaces-list"
    >
      <template #rows>
        <tr
          v-for="(namespace, i) in namespaces"
          :key="namespace.tenant_id || i"
        >
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
            <router-link
              :to="{ name: 'userDetails', params: { id: namespace.owner } }"
              class="unstyled-link text-decoration-underline cursor-pointer"
            >
              {{ getOwnerLabel(namespace) }}
            </router-link>
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
                  @click="goToNamespace(namespace.tenant_id)"
                  @keypress.enter="goToNamespace(namespace.tenant_id)"
                />
              </template>
              <span>Details</span>
            </v-tooltip>

            <v-tooltip
              bottom
              anchor="bottom"
            >
              <template #activator="{ props }">
                <v-icon
                  tag="button"
                  dark
                  v-bind="props"
                  tabindex="0"
                  aria-label="Edit Namespace"
                  data-test="namespace-edit-dialog-btn"
                  icon="mdi-pencil"
                  @click="openEditNamespace(namespace)"
                />
              </template>
              <span>Edit</span>
            </v-tooltip>

            <v-tooltip
              bottom
              anchor="bottom"
            >
              <template #activator="{ props }">
                <v-icon
                  tag="button"
                  dark
                  v-bind="props"
                  tabindex="0"
                  aria-label="Delete Namespace"
                  data-test="namespace-delete-dialog-btn"
                  icon="mdi-delete"
                  @click="openDeleteNamespace(namespace)"
                />
              </template>
              <span>Delete</span>
            </v-tooltip>
          </td>
        </tr>
      </template>
    </DataTable>

    <NamespaceEdit
      v-if="selectedNamespace"
      :key="selectedNamespace.tenant_id"
      v-model="namespaceEdit"
      :namespace="selectedNamespace"
      @update="fetchNamespaces"
    />

    <NamespaceDelete
      v-if="selectedNamespace"
      v-model="namespaceDelete"
      :tenant="selectedNamespace.tenant_id"
      :name="selectedNamespace.name"
      @update="fetchNamespaces"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import useNamespacesStore from "@admin/store/modules/namespaces";
import { IAdminNamespace } from "@admin/interfaces/INamespace";
import useSnackbar from "@/helpers/snackbar";
import DataTable from "@/components/Tables/DataTable.vue";
import NamespaceEdit from "@admin/components/Namespace/NamespaceEdit.vue";
import NamespaceDelete from "@admin/components/Namespace/NamespaceDelete.vue";
import handleError from "@/utils/handleError";

const snackbar = useSnackbar();
const namespacesStore = useNamespacesStore();
const namespaces = computed(() => namespacesStore.namespaces);
const namespaceCount = computed(() => namespacesStore.namespaceCount);
const router = useRouter();

const namespaceEdit = ref(false);
const namespaceDelete = ref(false);
const selectedNamespace = ref<IAdminNamespace | null>(null);

const loading = ref(false);
const page = ref(1);
const itemsPerPage = ref(10);

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
    text: "Actions",
    value: "actions",
  },
]);

const fetchNamespaces = async () => {
  try {
    loading.value = true;
    await namespacesStore.fetchNamespaceList({
      page: page.value,
      perPage: itemsPerPage.value,
    });
  } catch (error) {
    handleError(error);
    snackbar.showError("Failed to fetch namespaces.");
  } finally {
    loading.value = false;
  }
};

const sumDevicesCount = (namespace: IAdminNamespace) => {
  const {
    devices_accepted_count: acceptedCount,
    devices_pending_count: pendingCount,
    devices_rejected_count: rejectedCount,
  } = namespace;
  return (acceptedCount + pendingCount + rejectedCount) || 0;
};

const getOwnerLabel = (namespace: IAdminNamespace) => {
  const owner = namespace.members?.find(
    (member) => member.id === namespace.owner,
  );

  return owner?.email || namespace.owner || "";
};

const goToNamespace = async (tenantId: string) => {
  await router.push({ name: "namespaceDetails", params: { id: tenantId } });
};

const openEditNamespace = (ns: IAdminNamespace) => {
  selectedNamespace.value = ns;
  namespaceEdit.value = true;
};

const openDeleteNamespace = (ns: IAdminNamespace) => {
  selectedNamespace.value = ns;
  namespaceDelete.value = true;
};

watch([itemsPerPage, page], async () => {
  await fetchNamespaces();
});

onMounted(async () => {
  await fetchNamespaces();
});
</script>

<style scoped>
.unstyled-link {
  all: unset;
}
.unstyled-link:focus {
  outline: none;
}
</style>
