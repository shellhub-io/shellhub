<template>
  <v-container>
    <v-row align="center" justify="center">
      <v-col sm="8">
        <v-row>
          <v-col>
            <h3>Tenant ID:</h3>
          </v-col>
          <v-spacer />
          <v-col>
            <v-card tile :elevation="0" class="bg-v-theme-surface">
              <v-chip>
                <v-tooltip location="top">
                  <template v-slot:activator="{ props }">
                    <span
                      v-bind="props"
                      @click="copyText(tenant)"
                      @keypress="copyText(tenant)"
                      class="hover-text"
                      data-test="tenant-id"
                    >
                      {{ tenant }}
                      <v-icon icon="mdi-content-copy" />
                    </span>
                  </template>
                  <span>Copy ID</span>
                </v-tooltip>
              </v-chip>
            </v-card>
          </v-col>
        </v-row>

        <v-divider class="mt-6" />
        <v-divider class="mb-6" />

        <div class="mt-6" data-test="editOperation-div">
          <NamespaceEdit data-test="NamespaceEdit-component" />

          <v-divider class="mt-6" />
          <v-divider class="mb-6" />
        </div>

        <div v-if="true" class="mt-6" data-test="securityOperation-div">
          <SettingSecurity :hasTenant="hasTenant()" />

          <v-divider />
          <v-divider />
        </div>

        <div class="mt-6" data-test="deleteOperation-div">
          <h3 class="mb-2">Danger Zone</h3>
          <v-alert
            v-if="billingInDebt"
            type="error"
            text="We regret to inform you that it
            is currently not possible to delete
            your namespace due to either an outstanding
            unpaid invoice or an active subscription." />
          <v-row class="mt-2 mb-2" v-if="isOwner">
            <v-col class="ml-3">
              <h4>Delete this namespace</h4>
              <div class="ml-2">
                <p>
                  After deleting a namespace, there is no going back. Be sure.
                </p>
              </div>
            </v-col>

            <v-col md="auto" class="ml-auto mb-4">
              <NamespaceDelete :tenant="tenant" @billing-in-debt="billingInDebt = true" />
            </v-col>
          </v-row>
          <v-row class="mt-2 mb-2" v-else>
            <v-col class="ml-3">
              <h4>Leave this namespace</h4>
              <div class="ml-2">
                <p>
                  After leaving a namespace, you will need to be invited again to access it.
                </p>
              </div>
            </v-col>

            <v-col md="auto" class="ml-auto mb-4">
              <NamespaceLeave :tenant="tenant" />
            </v-col>
          </v-row>
        </div>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import { onMounted, computed, ref } from "vue";
import axios, { AxiosError } from "axios";
import { useStore } from "../../store";
import NamespaceEdit from "../Namespace/NamespaceEdit.vue";
import SettingSecurity from "./SettingSecurity.vue";
import NamespaceDelete from "../Namespace/NamespaceDelete.vue";
import {
  INotificationsCopy,
  INotificationsError,
} from "../../interfaces/INotifications";
import handleError from "@/utils/handleError";
import NamespaceLeave from "../Namespace/NamespaceLeave.vue";

const store = useStore();
const namespace = computed(() => store.getters["namespaces/get"]);
const isOwner = computed(() => namespace.value.owner === localStorage.getItem("id"));
const tenant = computed(() => store.getters["auth/tenant"]);
const billingInDebt = ref(false);

const copyText = (value: string | undefined) => {
  if (value) {
    navigator.clipboard.writeText(value);
    store.dispatch(
      "snackbar/showSnackbarCopy",
      INotificationsCopy.tenantId,
    );
  }
};

const getNamespace = async () => {
  try {
    await store.dispatch("namespaces/get", tenant.value);
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response?.status === 403) {
        store.dispatch("snackbar/showSnackbarErrorAssociation");
      }
    } else {
      store.dispatch(
        "snackbar/showSnackbarErrorAction",
        INotificationsError.namespaceLoad,
      );
      handleError(error);
    }
  }
};

onMounted(async () => {
  if (tenant.value) {
    await getNamespace();
  }
});

const hasTenant = () => tenant.value !== "";

</script>

<style scoped>
.hover-text {
  cursor: pointer;
}

.hover-text:hover {
  text-decoration: underline;
}
</style>
