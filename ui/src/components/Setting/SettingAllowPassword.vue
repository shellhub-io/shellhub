<template>
  <v-switch
    v-model="isAllowPasswordEnabled"
    hide-details
    inset
    :disabled="!canUpdateAllowPassword"
    color="primary"
    data-test="allow-password-switch"
  />
</template>

<script setup lang="ts">
import { computed } from "vue";
import hasPermission from "@/utils/permission";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useNamespacesStore from "@/store/modules/namespaces";

const props = defineProps<{ tenantId: string }>();

const snackbar = useSnackbar();
const namespacesStore = useNamespacesStore();

const updateAllowPasswordStatus = async (isAllowed: boolean) => {
  try {
    await namespacesStore.editNamespace({
      tenant_id: props.tenantId,
      settings: {
        allow_password: isAllowed,
      },
    });
    snackbar.showSuccess(`Password authentication was successfully ${isAllowed ? "allowed" : "disallowed"}.`);
  } catch (error: unknown) {
    snackbar.showError("Failed to update password authentication status.");
    handleError(error);
  }
};

const isAllowPasswordEnabled = computed({
  get: () => namespacesStore.currentNamespace.settings?.allow_password ?? true,
  set: (isAllowed: boolean) => {
    void updateAllowPasswordStatus(isAllowed);
  },
});

const canUpdateAllowPassword = hasPermission("namespace:updateAllowPassword");
</script>
