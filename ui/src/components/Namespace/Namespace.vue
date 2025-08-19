<template>
  <NamespaceAdd v-model="isAddNamespaceDialogVisible" />

  <v-select
    :menu-props="{ closeOnContentClick: true }"
    v-model="selectedNamespace"
    label="Active Namespace"
    variant="outlined"
    item-title="name"
    item-value="url"
    :items="namespaceList"
    :hide-details="true"
    class="mt-2"
  >
    <template #prepend-inner>
      <v-chip label color="primary" class="text-uppercase">{{ firstNamespaceLetter }}</v-chip>
    </template>
    <template #prepend-item>
      <v-list-subheader>
        All Namespaces
      </v-list-subheader>
    </template>
    <template #item="{ item }">
      <v-list-item @click="changeNamespace((item.raw as NamespaceItem).tenant_id)" title="">
        <v-chip label color="primary" class="text-uppercase mr-2">{{ (item.raw as NamespaceItem).name.charAt(0) }}</v-chip>
        <span>{{ (item.raw as NamespaceItem).name }}</span>
      </v-list-item>
    </template>

    <template #append-item>
      <v-divider />
      <v-list-item class="mt-2 mb-0">
        <v-btn
          variant="flat"
          prepend-icon="mdi-plus-box"
          color="primary"
          class="ma-0"
          block
          @click="isAddNamespaceDialogVisible = true"
        >New Namespace
        </v-btn>
      </v-list-item>
    </template>
  </v-select>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import axios, { AxiosError } from "axios";
import NamespaceAdd from "./NamespaceAdd.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useNamespacesStore from "@/store/modules/namespaces";

interface NamespaceItem {
  tenant_id: string;
  name: string;
}

defineOptions({
  inheritAttrs: false,
});

const snackbar = useSnackbar();
const namespacesStore = useNamespacesStore();
const namespaceList = computed(() => namespacesStore.namespaceList);
const selectedNamespace = computed(() => namespacesStore.currentNamespace);
const tenant = computed(() => localStorage.getItem("tenant") as string);
const firstNamespaceLetter = computed(() => (selectedNamespace.value.name ?? "").charAt(0));
const isAddNamespaceDialogVisible = ref(false);

// Change the current namespace
const changeNamespace = async (tenantId: string) => {
  try {
    await namespacesStore.switchNamespace(tenantId);
    window.location.reload();
  } catch (error: unknown) {
    snackbar.showError("An error occurred while switching namespaces.");
    handleError(error);
  }
};

// Fetch the current namespace
const fetchNamespace = async () => {
  try {
    await namespacesStore.fetchNamespace(tenant.value);
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      switch (true) {
        case axiosError.response?.status === 404: {
          // detects namespace inserted
          const namespace = namespacesStore.namespaceList[0];
          if (tenant.value === "" && namespace !== undefined) {
            changeNamespace(namespace.tenant_id);
          }
          break;
        }
        case axiosError.response?.status === 500 && tenant.value === null: {
          break;
        }
        default: {
          snackbar.showError("An error occurred while loading the namespace.");
          handleError(error);
        }
      }
    }
  }
};

onMounted(async () => {
  await fetchNamespace();
});
</script>
