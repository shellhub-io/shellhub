<template>
  <DataTable
    v-model:itemsPerPage="itemsPerPage"
    v-model:page="page"
    :headers
    :items="tags"
    :nextPage="next"
    :previousPage="prev"
    :loading
    :totalCount="numberTags"
    :itemsPerPageOptions="[10, 20, 50, 100]"
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
                      :has-authorization="hasAuthorizationEdit()"
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
                      :tag-name="item.name"
                      :has-authorization="hasAuthorizationRemove()"
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
import { actions, authorizer } from "../../authorizer";
import hasPermission from "../../utils/permission";
import DataTable from "../DataTable.vue";
import TagRemove from "./TagRemove.vue";
import TagEdit from "./TagEdit.vue";
import handleError from "@/utils/handleError";
import useTagsStore from "@/store/modules/tags";
import useSnackbar from "@/helpers/snackbar";
import useAuthStore from "@/store/modules/auth";

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

const tagsStore = useTagsStore();
const authStore = useAuthStore();
const snackbar = useSnackbar();
const loading = ref(false);
const itemsPerPage = ref(10);
const page = ref<number>(1);
const tags = computed(() => tagsStore.list);
const tenant = computed(() => localStorage.getItem("tenant"));
const numberTags = computed<number>(
  () => tagsStore.getNumberTags,
);

const getTags = async (perPage: number, page: number): Promise<void> => {
  if (!tenant.value) return;

  loading.value = true;

  try {
    await tagsStore.fetch({
      tenant: tenant.value,
      filter: tagsStore.getFilter || "",
      perPage,
      page,
    });

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
  if (page.value > 1) await getTags(itemsPerPage.value, page.value--);
};

const changeItemsPerPage = async (newItemsPerPage: number) => {
  itemsPerPage.value = newItemsPerPage;
};

watch(itemsPerPage, async (newItemsPerPage) => {
  await getTags(newItemsPerPage, page.value);
});

const hasAuthorizationEdit = () => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.tag.edit);
};

const hasAuthorizationRemove = () => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.tag.remove);
};

onMounted(() => {
  refresh();
});

defineExpose({ refresh });
</script>
