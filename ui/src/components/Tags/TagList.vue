<template>
  <DataTable
    :headers="headers"
    :items="tags"
    :itemsPerPage="itemsPerPage"
    :nextPage="next"
    :previousPage="prev"
    :loading="loading"
    :actualPage="page"
    :totalCount="numberTags"
    :comboboxOptions="[10, 20, 50, 100]"
    @changeItemsPerPage="changeItemsPerPage"
    @clickNextPage="next"
    @clickPreviousPage="prev"
    data-test="tag-list"
  >
    <template v-slot:rows>
      <tr v-for="(item, i) in tags" :key="i">
        <td class="text-center" data-test="tag-name"> {{ item.name }}</td>
        <td class="text-center">
          <v-menu location="bottom" scrim eager>
            <template v-slot:activator="{ props }">
              <v-btn
                v-bind="props"
                variant="plain"
                class="border rounded bg-v-theme-background"
                density="comfortable"
                size="default"
                icon="mdi-format-list-bulleted"
                data-test="tag-list-actions"
              />
            </template>
            <v-list class="bg-v-theme-surface" lines="two" density="compact">
              <v-tooltip location="bottom" class="text-center" :disabled="hasAuthorizationEdit()">
                <template v-slot:activator="{ props }">
                  <div v-bind="props">
                    <TagEdit
                      :tag-name="item.name"
                      :not-has-authorization="!hasAuthorizationEdit()"
                      @update="refresh()"
                    />
                  </div>
                </template>
                <span> You don't have this kind of authorization. </span>
              </v-tooltip>

              <v-tooltip location="bottom" class="text-center" :disabled="hasAuthorizationRemove()">
                <template v-slot:activator="{ props }">
                  <div v-bind="props">
                    <TagRemove
                      :tag="item.name"
                      :not-has-authorization="!hasAuthorizationRemove()"
                      @update="refresh()"
                    />
                  </div>
                </template>
                <span> You don't have this kind of authorization. </span>
              </v-tooltip>
            </v-list>
          </v-menu>
        </td>
      </tr>
    </template>
  </DataTable>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useStore } from "../../store";
import { FetchTagsParams } from "../../interfaces/ITags";
import { actions, authorizer } from "../../authorizer";
import hasPermission from "../../utils/permission";
import DataTable from "../DataTable.vue";
import TagRemove from "./TagRemove.vue";
import TagEdit from "./TagEdit.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";

const headers = ref([
  {
    text: "Name",
    value: "name",
  },
  {
    text: "Actions",
    value: "actions",
  },
]);

const store = useStore();
const snackbar = useSnackbar();
const loading = ref(false);
const itemsPerPage = ref(10);
const page = ref<number>(1);
const tags = computed(() => store.getters["tags/list"]);
const tenant = computed(() => localStorage.getItem("tenant"));
const numberTags = computed<number>(
  () => store.getters["tags/getNumberTags"],
);

const getTags = async (perPage: number, page: number): Promise<void> => {
  if (!tenant.value) return;

  loading.value = true;

  try {
    await store.dispatch("tags/fetch", {
      tenant: tenant.value,
      filter: store.getters["tags/getFilter"],
      perPage,
      page,
    } as FetchTagsParams);

    loading.value = false;
  } catch (error: unknown) {
    snackbar.showError("Failed to load tags.");
    handleError(error);
  }
};

const refresh = async () => {
  await getTags(itemsPerPage.value, page.value);
};

const next = async () => {
  await getTags(itemsPerPage.value, page.value++);
};

const prev = async () => {
  try {
    if (page.value > 1) await getTags(itemsPerPage.value, page.value--);
  } catch (error) {
    store.dispatch("snackbar/setSnackbarErrorDefault");
  }
};

const changeItemsPerPage = async (newItemsPerPage: number) => {
  itemsPerPage.value = newItemsPerPage;
};

watch(itemsPerPage, async (newItemsPerPage) => {
  await getTags(newItemsPerPage, page.value);
});

const hasAuthorizationEdit = () => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(authorizer.role[role], actions.tag.edit);
  }
  return false;
};

const hasAuthorizationRemove = () => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(authorizer.role[role], actions.tag.remove);
  }
  return false;
};

onMounted(() => {
  refresh();
});

defineExpose({ refresh });
</script>
