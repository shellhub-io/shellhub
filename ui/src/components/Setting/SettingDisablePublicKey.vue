<template>
  <v-switch
    v-model="isDisablePublicKeyEnabled"
    hide-details
    inset
    :disabled="!canUpdateDisablePublicKey"
    color="primary"
    data-test="disable-public-key-switch"
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

const updateDisablePublicKeyStatus = async (isDisabled: boolean) => {
  try {
    await namespacesStore.editNamespace({
      tenant_id: props.tenantId,
      settings: {
        disable_public_key: isDisabled,
      },
    });
    snackbar.showSuccess(`Public key authentication was successfully ${isDisabled ? "disabled" : "enabled"}.`);
  } catch (error: unknown) {
    snackbar.showError("Failed to update public key authentication status.");
    handleError(error);
  }
};

const isDisablePublicKeyEnabled = computed({
  get: () => namespacesStore.currentNamespace.settings?.disable_public_key || false,
  set: (isDisabled: boolean) => {
    void updateDisablePublicKeyStatus(isDisabled);
  },
});

const canUpdateDisablePublicKey = hasPermission("namespace:updateDisablePublicKey");
</script>