<template>
  <BaseDialog v-model="showDialog">
    <v-card data-test="namespace-delete-card" class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-4 bg-primary">
        Namespace Deletion
      </v-card-title>

      <v-card-text class="mt-4 mb-3 pb-1">
        <div
          v-if="isBillingActive"
          data-test="content-subscription-text"
        >
          <p class="mb-2">
            To ensure the integrity of your namespace,
            we have implemented a restriction that prevents its deletion while you have an active subscription or an unpaid invoice.
          </p>
          <p class="mb-2">
            Kindly note that in order to proceed with the deletion of your namespace,
            please ensure that there are no active subscriptions associated with it, and all outstanding invoices are settled.
          </p>
        </div>

        <p data-test="content-text" v-else>
          This action cannot be undone. This will permanently delete the
          <b> {{ displayOnlyTenCharacters(name) }} </b> and its related data.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" data-test="close-btn" @click="showDialog = false">
          Close
        </v-btn>

        <v-btn
          color="red darken-1"
          variant="text"
          data-test="remove-btn"
          @click="remove()"
          :disabled="isBillingActive || !hasAuthorization"
        >
          Remove
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import axios, { AxiosError } from "axios";
import hasPermission from "@/utils/permission";
import { actions, authorizer } from "@/authorizer";
import { displayOnlyTenCharacters } from "@/utils/string";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
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
const showDialog = defineModel({ default: false });
const { name } = namespacesStore.currentNamespace;
const tenant = computed(() => props.tenant);
const isBillingActive = computed(() => billingStore.isActive);
const hasAuthorization = computed(() => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.namespace.remove);
});

const remove = async () => {
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
  }
};

onMounted(async () => {
  if (hasAuthorization.value && envVariables.isCloud) {
    await billingStore.getSubscriptionInfo();
  }
});
</script>
