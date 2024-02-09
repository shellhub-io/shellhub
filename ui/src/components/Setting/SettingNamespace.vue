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

<script lang="ts">
import { defineComponent, onMounted, computed, ref } from "vue";
import axios, { AxiosError } from "axios";
import { envVariables } from "../../envVariables";
import { useStore } from "../../store";
import NamespaceEdit from "../Namespace/NamespaceEdit.vue";
import NamespaceMemberAdd from "../Namespace/NamespaceMemberAdd.vue";
import NamespaceMemberList from "../Namespace/NamespaceMemberList.vue";
import SettingSecurity from "./SettingSecurity.vue";
import NamespaceDelete from "../Namespace/NamespaceDelete.vue";
import {
  INotificationsCopy,
  INotificationsError,
} from "../../interfaces/INotifications";
import handleError from "@/utils/handleError";

export default defineComponent({
  setup() {
    const store = useStore();
    const namespace = computed(() => store.getters["namespaces/get"]);
    const tenant = computed(() => store.getters["auth/tenant"]);
    const isEnterprise = computed(() => envVariables.isEnterprise);
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

    onMounted(async () => {
      if (tenant.value) {
        await getNamespace();
      }
    });

    const hasTenant = () => tenant.value !== "";

    return {
      tenant,
      namespace,
      copyText,
      refresh,
      isEnterprise,
      hasTenant,
      billingInDebt,
    };
  },
  components: {
    NamespaceEdit,
    NamespaceMemberAdd,
    NamespaceMemberList,
    SettingSecurity,
    NamespaceDelete,
  },
});
</script>

<style scoped>
.hover-text {
  cursor: pointer;
}

.hover-text:hover {
  text-decoration: underline;
}
</style>
