<template>
  <div
    class="d-flex flex-column justify-space-between align-center flex-sm-row mb-2"
    data-test="public-keys-title"
  >
    <h1>Public Keys</h1>

    <v-spacer />
    <v-spacer />

    <PublicKeyAdd @update="refresh" />
  </div>

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
import { useStore } from "../store";
import NoItemsMessage from "../components/NoItemsMessage.vue";
import PublicKeyAdd from "../components/PublicKeys/PublicKeyAdd.vue";
import PublicKeysList from "../components/PublicKeys/PublicKeysList.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";

const store = useStore();
const snackbar = useSnackbar();
const hasPublicKey = computed(
  () => store.getters["publicKeys/getNumberPublicKeys"] > 0,
);

const refresh = async () => {
  try {
    await store.dispatch("publicKeys/refresh");
  } catch (error: unknown) {
    snackbar.showError("Failed to load the public keys list.");
    handleError(error);
  }
};

onMounted(async () => {
  store.dispatch("publicKeys/resetPagePerpage");
  await refresh();
});

defineExpose({ refresh });
</script>
