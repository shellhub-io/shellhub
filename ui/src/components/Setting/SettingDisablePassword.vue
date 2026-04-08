<template>
  <v-switch
    v-model="isDisablePasswordEnabled"
    hide-details
    inset
    :disabled="!canUpdateDisablePassword"
    color="primary"
    data-test="disable-password-switch"
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

const updateDisablePasswordStatus = async (isDisabled: boolean) => {
  try {
    await namespacesStore.editNamespace({
      tenant_id: props.tenantId,
      settings: {
        disable_password: isDisabled,
      },
    });
    snackbar.showSuccess(`Password authentication was successfully ${isDisabled ? "disabled" : "enabled"}.`);
  } catch (error: unknown) {
    snackbar.showError("Failed to update password authentication status.");
    handleError(error);
  }
};

const isDisablePasswordEnabled = computed({
  get: () => namespacesStore.currentNamespace.settings?.disable_password || false,
  set: (isDisabled: boolean) => {
    void updateDisablePasswordStatus(isDisabled);
  },
});

const canUpdateDisablePassword = hasPermission("namespace:updateDisablePassword");
</script>
