<template>
  <PageHeader
    icon="mdi-key"
    title="Public Keys"
    overline="SSH Authentication"
    description="Manage SSH public keys for secure, password-free authentication to your devices. SSH keys are more secure than passwords."
    icon-color="primary"
    data-test="public-keys-title"
  >
    <template #actions>
      <PublicKeyAdd @update="refresh" />
    </template>
  </PageHeader>

  <div data-test="public-keys-components">
    <PublicKeysList v-if="hasPublicKey" />

    <NoItemsMessage
      v-else
      item="Public Keys"
      icon="mdi-key"
      data-test="no-items-message-component"
    >
      <template #content>
        <p>You can connect to your devices using password-based logins, but we strongly recommend using SSH key pairs instead.</p>
        <p>SSH keys are more secure than passwords and can help you log in without having to remember long passwords.</p>
      </template>
      <template #action>
        <PublicKeyAdd @update="refresh" />
      </template>
    </NoItemsMessage>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import NoItemsMessage from "../components/NoItemsMessage.vue";
import PublicKeyAdd from "../components/PublicKeys/PublicKeyAdd.vue";
import PublicKeysList from "../components/PublicKeys/PublicKeysList.vue";
import PageHeader from "../components/PageHeader.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import usePublicKeysStore from "@/store/modules/public_keys";

const publicKeysStore = usePublicKeysStore();
const snackbar = useSnackbar();
const hasPublicKey = computed(() => publicKeysStore.publicKeyCount > 0);

const refresh = async () => {
  try {
    await publicKeysStore.fetchPublicKeyList();
  } catch (error: unknown) {
    snackbar.showError("Failed to load the public keys list.");
    handleError(error);
  }
};

onMounted(async () => { await refresh(); });
</script>
