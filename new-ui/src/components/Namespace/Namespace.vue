<template>
  <v-list v-if="hasNamespace">
    <v-list-group v-model="listing">
      <template v-slot:activator="{ props }">
        <v-list-item v-bind="props" class="text-primary icon-primary">
          <v-list-item-title> {{ namespace.name }}</v-list-item-title>
        </v-list-item>
      </template>

      <NamespaceList data-test="namespaceList-component" />

      <v-list-item v-if="isEnterprise">
        <NamespaceAdd
          data-test="namespaceAdd-component"
          @update="getNamespaces"
        />
      </v-list-item>
    </v-list-group>
  </v-list>
  <div v-else>
    <NamespaceAdd data-test="namespaceAdd-component" @update="getNamespaces" />
  </div>
</template>

<script lang="ts">
import { useStore } from "../../store";
import { defineComponent, ref, computed, watch, onMounted } from "vue";
import { envVariables } from "../../envVariables";
import { INotificationsError } from "../../interfaces/INotifications";
import NamespaceList from "./NamespaceList.vue";
import NamespaceAdd from "./NamespaceAdd.vue";

export default defineComponent({
  inheritAttrs: false,
  setup() {
    const store = useStore();
    const inANamespace = ref(false);
    const listing = ref(false);
    const namespace = computed(() => store.getters["namespaces/get"]);
    const hasNamespace = computed(
      () => store.getters["namespaces/getNumberNamespaces"] !== 0
    );
    const tenant = computed(() => localStorage.getItem("tenant"));
    const isEnterprise = computed(() => envVariables.isEnterprise);
    onMounted(async () => {
      await getNamespaces();
      if (inANamespace.value) {
        await getNamespace();
      }
    });
    watch(hasNamespace, (status) => {
      inANamespace.value = status;
      getNamespace();
    });
    watch(listing, (val) => {
      if (val) {
        getNamespaces();
      }
    });
    const getNamespace = async () => {
      if (!store.getters['auth/isLoggedIn']) return;

      try {
        await store.dispatch("namespaces/get", tenant.value);
      } catch (error: any) {
        switch (true) {
          case error.response.status === 404: {
            // detects namespace inserted
            const namespaceFind = store.getters["namespaces/list"][0];
            if (tenant.value === "" && namespaceFind !== undefined) {
              switchIn(namespaceFind.tenant_id);
            }
            break;
          }
          case error.response.status === 500 && tenant === null: {
            break;
          }
          default: {
            store.dispatch(
              "snackbar/showSnackbarErrorLoading",
              INotificationsError.namespaceLoad
            );
          }
        }
      }
    };
    const getNamespaces = async () => {
      try {
        if (!store.getters['auth/isLoggedIn']) return;
        await store.dispatch("namespaces/fetch", {
          page: 1,
          perPage: 30,
        });
      } catch (error: any) {
        switch (true) {
          case !inANamespace.value && error.response.status === 403: {
            // dialog pops
            break;
          }
          case error.response.status === 403: {
            store.dispatch("snackbar/showSnackbarErrorAssociation");
            break;
          }
          default: {
            store.dispatch(
              "snackbar/showSnackbarErrorLoading",
              INotificationsError.namespaceList
            );
          }
        }
      }
    };
    const switchIn = async (tenantId: string) => {
      try {
        await store.dispatch("namespaces/switchNamespace", {
          tenant_id: tenantId,
        });
        window.location.reload();
      } catch {
        store.dispatch(
          "snackbar/showSnackbarErrorLoading",
          INotificationsError.namespaceSwitch
        );
      }
    };
    return {
      inANamespace,
      hasNamespace,
      tenant,
      listing,
      switchIn,
      namespace,
      isEnterprise,
      getNamespaces,
    };
  },
  components: { NamespaceList, NamespaceAdd },
});
</script>
