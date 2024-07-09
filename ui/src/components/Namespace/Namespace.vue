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
    class="mt-1"
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
      <v-list-item @click="changeNamespace(item.raw.tenant_id)" title="">
        <v-chip label color="primary" class="text-uppercase mr-2">{{ item.raw.name.charAt(0) }}</v-chip>
        <span>{{ item.raw.name }}</span>
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
import { useStore } from "@/store";
import { INotificationsError } from "@/interfaces/INotifications";
import NamespaceAdd from "./NamespaceAdd.vue";
import handleError from "@/utils/handleError";

const store = useStore();
const namespaceList = computed(() => store.getters["namespaces/list"]);
const selectedNamespace = computed(() => store.getters["namespaces/get"]);
const tenant = computed(() => localStorage.getItem("tenant"));
const firstNamespaceLetter = computed(() => (store.getters["namespaces/get"].name ?? "").charAt(0));
const isAddNamespaceDialogVisible = ref(false);

// Change the current namespace
const changeNamespace = async (tenantId: string) => {
  try {
    await store.dispatch("namespaces/switchNamespace", {
      tenant_id: tenantId,
    });
    window.location.reload();
  } catch (error: unknown) {
    store.dispatch(
      "snackbar/showSnackbarErrorLoading",
      INotificationsError.namespaceSwitch,
    );
    handleError(error);
  }
};

// Fetch the current namespace
const fetchNamespace = async () => {
  try {
    await store.dispatch("namespaces/get", tenant.value);
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      switch (true) {
        case axiosError.response?.status === 404: {
          // detects namespace inserted
          const namespaceFind = store.getters["namespaces/list"][0];
          if (tenant.value === "" && namespaceFind !== undefined) {
            changeNamespace(namespaceFind.tenant_id);
          }
          break;
        }
        case axiosError.response?.status === 500 && tenant.value === null: {
          break;
        }
        default: {
          store.dispatch(
            "snackbar/showSnackbarErrorLoading",
            INotificationsError.namespaceLoad,
          );
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
