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
            <v-card tile :elevation="0">
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
          <NamespaceRename data-test="namespaceRename-component" />

          <v-divider />
          <v-divider />
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
              <NamespaceDelete :nsTenant="tenant" />
            </v-col>
          </v-row>
        </div>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import { defineComponent, onMounted, ref, computed } from "vue";
import { envVariables } from "../../envVariables";
import { useStore } from "../../store";
import NamespaceRename from "../Namespace/NamespaceRename.vue";
import NamespaceMemberAdd from "../Namespace/NamespaceMemberAdd.vue";
import NamespaceMemberList from "../Namespace/NamespaceMemberList.vue";
import SettingSecurity from "./SettingSecurity.vue";
import NamespaceDelete from "../Namespace/NamespaceDelete.vue";
import {
  INotificationsCopy,
  INotificationsError,
} from "../../interfaces/INotifications";

export default defineComponent({
  setup() {
    const store = useStore();
    const namespace = computed(() => store.getters["namespaces/get"]);
    const tenant = computed(() => store.getters["auth/tenant"]);
    const isEnterprise = computed(() => envVariables.isEnterprise);

    const copyText = (value: string | undefined) => {
      if (value) {
        navigator.clipboard.writeText(value);
        store.dispatch(
          "snackbar/showSnackbarCopy",
          INotificationsCopy.tenantId
        );
      }
    };

    const refresh = () => {
      getNamespace();
    };

    const getNamespace = async () => {
      try {
        await store.dispatch("namespaces/get", tenant.value);
      } catch (error: any) {
        if (error.response.status === 403) {
          store.dispatch("snackbar/showSnackbarErrorAssociation");
        } else {
          store.dispatch(
            "snackbar/showSnackbarErrorAction",
            INotificationsError.namespaceLoad
          );
        }
      }
    };

    onMounted(async () => {
      if (tenant.value) {
        await getNamespace();
      }
    });

    const hasTenant = () => {
      return tenant.value !== "";
    };

    return {
      tenant,
      namespace,
      copyText,
      refresh,
      isEnterprise,
      hasTenant,
    };
  },
  components: {
    NamespaceRename,
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
