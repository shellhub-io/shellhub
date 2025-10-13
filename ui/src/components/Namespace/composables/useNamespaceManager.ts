import { computed } from "vue";
import axios, { AxiosError } from "axios";
import useNamespacesStore from "@/store/modules/namespaces";
import useSnackbar from "@/helpers/snackbar";
import handleError from "@/utils/handleError";

export default function useNamespaceManager() {
  const namespacesStore = useNamespacesStore();
  const snackbar = useSnackbar();

  const currentNamespace = computed(() => namespacesStore.currentNamespace);
  const namespaceList = computed(() => namespacesStore.namespaceList);
  const hasNamespaces = computed(() => namespaceList.value.length > 0);
  const currentTenantId = computed(() => localStorage.getItem("tenant") || "");

  const switchNamespace = async (tenantId: string) => {
    if (tenantId === currentNamespace.value.tenant_id) {
      return;
    }

    try {
      await namespacesStore.switchNamespace(tenantId);
      window.location.reload();
    } catch (error: unknown) {
      snackbar.showError("Failed to switch namespace");
      handleError(error);
    }
  };

  const loadCurrentNamespace = async () => {
    try {
      await namespacesStore.fetchNamespace(currentTenantId.value);
    } catch (error: unknown) {
      if (!axios.isAxiosError(error)) {
        snackbar.showError("Failed to load namespace");
        handleError(error);
        return;
      }

      const axiosError = error as AxiosError;

      // Namespace not found, try to switch to first available
      if (axiosError.response?.status === 404) {
        const firstNamespace = namespaceList.value[0];
        if (currentTenantId.value === "" && firstNamespace) {
          await switchNamespace(firstNamespace.tenant_id);
        }
        return;
      }

      // Server error with no tenant - ignore
      if (axiosError.response?.status === 500 && !currentTenantId.value) {
        return;
      }

      snackbar.showError("Failed to load namespace");
      handleError(error);
    }
  };

  return {
    currentNamespace,
    namespaceList,
    hasNamespaces,
    switchNamespace,
    loadCurrentNamespace,
  };
}
