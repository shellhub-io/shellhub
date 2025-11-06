<template>
  <MessageDialog
    v-model="showDialog"
    title="Namespace Deletion"
    icon="mdi-delete-alert"
    icon-color="error"
    :confirm-text="isBillingActive || !canDeleteNamespace ? '' : 'Remove'"
    confirm-color="error"
    :confirm-disabled="isBillingActive || !canDeleteNamespace"
    :confirm-loading="isLoading"
    cancel-text="Close"
    confirm-data-test="remove-btn"
    cancel-data-test="close-btn"
    @close="showDialog = false"
    @confirm="remove"
    @cancel="showDialog = false"
  >
    <div
      v-if="isBillingActive"
      data-test="content-subscription-text"
    >
      <p class="mb-2">
        To ensure the integrity of your namespace, we have implemented a
        restriction that prevents its deletion while you have an active
        subscription or an unpaid invoice.
      </p>
      <p class="mb-2">
        Kindly note that in order to proceed with the deletion of your
        namespace, please ensure that there are no active subscriptions
        associated with it, and all outstanding invoices are settled.
      </p>
    </div>

    <p
      v-else
      data-test="content-text"
    >
      This action cannot be undone. This will permanently delete
      <strong>{{ displayOnlyTenCharacters(name) }}</strong> and its related
      data.
    </p>
  </MessageDialog>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import axios, { AxiosError } from "axios";
import hasPermission from "@/utils/permission";
import { displayOnlyTenCharacters } from "@/utils/string";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
import useAuthStore from "@/store/modules/auth";
import useBillingStore from "@/store/modules/billing";
import useNamespacesStore from "@/store/modules/namespaces";
import { envVariables } from "@/envVariables";

const props = defineProps<{ tenant: string }>();
const emit = defineEmits(["billing-in-debt"]);

const authStore = useAuthStore();
const billingStore = useBillingStore();
const namespacesStore = useNamespacesStore();
const snackbar = useSnackbar();
const router = useRouter();
const showDialog = defineModel<boolean>({ required: true });
const isLoading = ref(false);
const { name } = namespacesStore.currentNamespace;
const tenant = computed(() => props.tenant);
const isBillingActive = computed(() => billingStore.isActive);
const canDeleteNamespace = hasPermission("namespace:delete");

const remove = async () => {
  isLoading.value = true;
  try {
    await namespacesStore.deleteNamespace(tenant.value);
    snackbar.showSuccess("Namespace deleted successfully.");
    authStore.logout();
    await router.push({ name: "Login" });
    showDialog.value = false;
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      switch (axiosError.response?.status) {
        case 402:
          emit("billing-in-debt");
          break;
        default:
          break;
      }
    }
    snackbar.showError("An error occurred while deleting the namespace.");
    handleError(error);
  } finally {
    isLoading.value = false;
  }
};

onMounted(async () => {
  if (canDeleteNamespace && envVariables.isCloud) {
    await billingStore.getSubscriptionInfo();
  }
});
</script>
