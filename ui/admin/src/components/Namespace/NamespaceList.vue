<template>
  <div>
    <DataTable
      :headers="headers"
      :items="namespaces"
      :itemsPerPage="itemsPerPage"
      :loading="loading"
      :actualPage="page"
      :total-count="numberOfNamespaces"
      @changeItemsPerPage="changeItemsPerPage"
      @clickNextPage="next"
      @clickPreviousPage="prev"
      data-test="namespaces-list"
    >
      <template v-slot:rows>
        <tr v-for="(namespace, i) in namespaces" :key="i">
          <td>
            {{ namespace.name }}
          </td>
          <td>
            {{ namespace.devices_count || 0 }}
          </td>
          <td>
            {{ namespace.tenant_id }}
          </td>
          <td>
            {{ namespace.owner }}
          </td>
          <td>
            <div v-if="namespace.settings">
              {{ namespace.settings.session_record }}
            </div>
          </td>
          <td>
            <v-tooltip bottom anchor="bottom">
              <template v-slot:activator="{ props }">
                <v-icon
                  tag="a"
                  dark
                  v-bind="props"
                  @click="goToNamespace(namespace.tenant_id)"
                  @keypress.enter="goToNamespace(namespace.tenant_id)"
                  tabindex="0"
                >mdi-information
                </v-icon>
              </template>
              <span>Details</span>
            </v-tooltip>

            <NamespaceEdit :namespace="namespace" />
          </td>
        </tr>
      </template>
    </DataTable>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { INotificationsError } from "../../interfaces/INotifications";
import { useStore } from "../../store";
import DataTable from "../DataTable.vue";
import NamespaceEdit from "./NamespaceEdit.vue";

export default defineComponent({
  setup() {
    const store = useStore();
    const router = useRouter();
    const loading = ref(false);
    const page = ref(1);
    const itemsPerPage = ref(10);
    const filter = ref("");

    onMounted(async () => {
      try {
        loading.value = true;
        await store.dispatch("namespaces/fetch", {
          perPage: itemsPerPage.value,
          page: page.value,
          filter: filter.value,
        });
      } catch {
        store.dispatch("snackbar/showSnackbarErrorAction", INotificationsError.namespaceList);
      } finally {
        loading.value = false;
      }
    });

    const namespaces = computed(() => store.getters["namespaces/list"]);

    const numberOfNamespaces = computed(() => store.getters["namespaces/numberOfNamespaces"]);

    const goToNamespace = (namespace: string) => {
      router.push({ name: "namespaceDetails", params: { id: namespace } });
    };

    const getNamespaces = async (perPagaeValue: number, pageValue: number) => {
      try {
        loading.value = true;
        const hasNamespaces = await store.dispatch("namespaces/fetch", {
          page: pageValue,
          perPage: perPagaeValue,
        });

        if (!hasNamespaces) page.value--;
      } catch {
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.namespaceList,
        );
      } finally {
        loading.value = false;
      }
    };

    const next = async () => {
      await getNamespaces(itemsPerPage.value, ++page.value);
    };

    const prev = async () => {
      try {
        if (page.value > 1) await getNamespaces(itemsPerPage.value, --page.value);
      } catch (error) {
        store.dispatch("snackbar/showSnackbarErrorDefault");
      }
    };

    const changeItemsPerPage = async (newItemsPerPage: number) => {
      itemsPerPage.value = newItemsPerPage;
    };

    watch(itemsPerPage, async () => {
      await getNamespaces(itemsPerPage.value, page.value);
    });

    return {
      headers: [
        {
          text: "Name",
          value: "name",
        },
        {
          text: "Devices",
          value: "devices",
        },
        {
          text: "Tenant ID",
          value: "tenant_id",
        },
        {
          text: "Owner",
          value: "owner",
        },
        {
          text: "Session Record",
          value: "settings",
        },
        {
          text: "Actions",
          value: "actions",
        },
      ],
      namespaces,
      numberOfNamespaces,
      loading,
      page,
      itemsPerPage,
      goToNamespace,
      next,
      prev,
      changeItemsPerPage,
    };
  },
  components: { DataTable, NamespaceEdit },
});
</script>
