<template>
  <MessageDialog
    v-model="showDialog"
    title="Namespace Leave"
    description="Are you sure you want to leave this namespace? If you leave, you will need an invitation to rejoin."
    icon="mdi-exit-to-app"
    icon-color="warning"
    confirm-text="Leave"
    confirm-color="error"
    :confirm-loading="isLoading"
    cancel-text="Close"
    confirm-data-test="leave-btn"
    cancel-data-test="close-btn"
    data-test="namespace-leave-dialog"
    @close="showDialog = false"
    @confirm="leave"
    @cancel="showDialog = false"
  />
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
import useNamespacesStore from "@/store/modules/namespaces";

const namespacesStore = useNamespacesStore();
const router = useRouter();
const snackbar = useSnackbar();
const showDialog = defineModel<boolean>({ required: true });
const isLoading = ref(false);
const tenant = computed(() => localStorage.getItem("tenant") as string);

const leave = async () => {
  isLoading.value = true;
  try {
    await namespacesStore.leaveNamespace(tenant.value);
    showDialog.value = false;
    snackbar.showSuccess("You have left the namespace.");
    router.go(0);
  } catch (error: unknown) {
    snackbar.showError("Failed to leave the namespace.");
    handleError(error);
  } finally {
    isLoading.value = false;
  }
};

defineExpose({ showDialog });
</script>
