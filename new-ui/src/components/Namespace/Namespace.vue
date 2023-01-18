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
          isSmall
          data-test="namespaceAdd-component"
          @update="getNamespaces"
        />
      </v-list-item>
    </v-list-group>
  </v-list>
  <div v-else>
    <NamespaceAdd
      enableSwitchIn
      isSmall
      data-test="namespaceAdd-component"
      @update="getNamespaces"
    />
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
    const isChecking = ref(false);
    const namespace = computed(() => store.getters["namespaces/get"]);
    const hasNamespace = computed(
      () => store.getters["namespaces/getNumberNamespaces"] !== 0
    );
    const openVersion = computed(() => !envVariables.isEnterprise);
    const tenant = computed(() => localStorage.getItem("tenant"));
    const isEnterprise = computed(() => envVariables.isEnterprise);
    onMounted(async () => {
      await getNamespaces();
      if (inANamespace.value) {
        await getNamespace();
      }
      if (Object.keys(namespace.value).length === 0 && openVersion.value) {
        isChecking.value = true;
        // Interval to check if the namespace has been added by cli
        setInterval(() => {
          checkNewNamespace();
        }, 3000);
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

    const checkNewNamespace = async () => {
      if (!store.getters["auth/isLoggedIn"]) return;

      await store.dispatch("namespaces/fetch", {
        page: 1,
        perPage: 10,
        fitler: "",
      });
      if (store.getters["namespaces/list"].length > 0) {
        switchIn(store.getters["namespaces/list"][0].tenant_id);
      }
    };

    const getNamespace = async () => {
      if (!store.getters["auth/isLoggedIn"]) return;
      if (isChecking.value) return;

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
            throw new Error(error);
          }
        }
      }
    };
    const getNamespaces = async () => {
      try {
        if (!store.getters["auth/isLoggedIn"]) return;
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
            throw new Error(error);
          }
          default: {
            store.dispatch(
              "snackbar/showSnackbarErrorLoading",
              INotificationsError.namespaceList
            );
            throw new Error(error);
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
      } catch (error: any) {
        store.dispatch(
          "snackbar/showSnackbarErrorLoading",
          INotificationsError.namespaceSwitch
        );
        throw new Error(error);
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
