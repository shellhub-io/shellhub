<template>
  <v-container>
    <v-row align="center" justify="center" class="mt-4">
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

        <div class="mt-6" data-test="userOperation-div">
          <v-row>
            <v-col>
              <h3>Members</h3>
            </v-col>

            <v-spacer />

            <v-col md="auto" class="ml-auto">
              <NamespaceMemberAdd @update="refresh" />
            </v-col>
          </v-row>

          <NamespaceMemberList :namespace="namespace" />
        </div>
        <v-divider class="mt-6" />
        <v-divider class="mb-6" />

        <div class="mt-6">
          <v-row>
            <v-col data-test="api-key-title">
              <h3>Api Keys</h3>
            </v-col>

            <v-col md="auto" class="ml-auto">
              <NamespaceGenerateApiKey @update="refreshApiKeys" data-test="api-key-generate" />
            </v-col>
          </v-row>

          <v-spacer />
          <v-row class="mt-2 mb-2">
            <v-col class="ml-3" data-test="api-key-text">
              Generate a Api Key for quick access to your ShellHub account.
            </v-col>
          </v-row>

          <v-row class="mt-2 mb-2">
            <v-col>
              <NamespaceApiKeyList ref="apiKeyList" data-test="api-key-list" />
            </v-col>
          </v-row>
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
          <v-row class="mt-2 mb-2">
            <v-col class="ml-3">
              <h4>Delete this namespace</h4>
              <div class="ml-2">
                <p>
                  After deleting a namespace, there is no going back. Be sure.
                </p>
              </div>
            </v-col>

            <v-col md="auto" class="ml-auto mb-4">
              <NamespaceDelete :nsTenant="tenant" @billing-in-debt="billingInDebt = true" />
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
import NamespaceMemberAdd from "../Namespace/NamespaceMemberAdd.vue";
import NamespaceMemberList from "../Namespace/NamespaceMemberList.vue";
import NamespaceGenerateApiKey from "../Namespace/NamespaceGenerateApiKey.vue";
import NamespaceApiKeyList from "../Namespace/NamespaceApiKeyList.vue";
import SettingSecurity from "./SettingSecurity.vue";
import NamespaceDelete from "../Namespace/NamespaceDelete.vue";
import {
  INotificationsCopy,
  INotificationsError,
} from "../../interfaces/INotifications";
import handleError from "@/utils/handleError";

const store = useStore();
const apiKeyList = ref();
const namespace = computed(() => store.getters["namespaces/get"]);
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
const refresh = () => {
  getNamespace();
};

const refreshApiKeys = () => {
  apiKeyList.value.refresh();
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
