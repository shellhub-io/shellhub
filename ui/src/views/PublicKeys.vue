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

    <BoxMessage
      v-if="showBoxMessage"
      typeMessage="publicKey"
      data-test="BoxMessagePublicKey-component"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useStore } from "../store";
import BoxMessage from "../components/Box/BoxMessage.vue";
import PublicKeyAdd from "../components/PublicKeys/PublicKeyAdd.vue";
import PublicKeysList from "../components/PublicKeys/PublicKeysList.vue";
import { INotificationsError } from "../interfaces/INotifications";
import handleError from "@/utils/handleError";

const store = useStore();
const show = ref(false);
const hasPublicKey = computed(
  () => store.getters["publicKeys/getNumberPublicKeys"] > 0,
);
const showBoxMessage = computed(() => !hasPublicKey.value && show.value);

const refresh = async () => {
  try {
    await store.dispatch("publicKeys/refresh");
  } catch (error: unknown) {
    store.dispatch(
      "snackbar/showSnackbarErrorLoading",
      INotificationsError.firewallRuleList,
    );
    handleError(error);
  }
};

onMounted(async () => {
  store.dispatch("box/setStatus", true);
  store.dispatch("publicKeys/resetPagePerpage");
  await refresh();
  show.value = true;
});

defineExpose({ refresh, showBoxMessage });
</script>
