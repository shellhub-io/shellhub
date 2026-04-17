<template>
  <v-switch
    v-model="isAllowPublicKeyEnabled"
    hide-details
    inset
    :disabled="!canUpdateAllowPublicKey"
    color="primary"
    data-test="allow-public-key-switch"
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

const updateAllowPublicKeyStatus = async (isAllowed: boolean) => {
  try {
    await namespacesStore.editNamespace({
      tenant_id: props.tenantId,
      settings: {
        allow_public_key: isAllowed,
      },
    });
    snackbar.showSuccess(`Public key authentication was successfully ${isAllowed ? "allowed" : "disallowed"}.`);
  } catch (error: unknown) {
    snackbar.showError("Failed to update public key authentication status.");
    handleError(error);
  }
};

const isAllowPublicKeyEnabled = computed({
  get: () => namespacesStore.currentNamespace.settings?.allow_public_key ?? true,
  set: (isAllowed: boolean) => {
    void updateAllowPublicKeyStatus(isAllowed);
  },
});

const canUpdateAllowPublicKey = hasPermission("namespace:updateAllowPublicKey");
</script>
