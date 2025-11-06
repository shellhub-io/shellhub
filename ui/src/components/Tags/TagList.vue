<template>
  <DataTable
    v-model:items-per-page="itemsPerPage"
    v-model:page="page"
    :headers
    :items="tags"
    :loading
    :total-count="numberTags"
    :items-per-page-options="[10, 20, 50, 100]"
    data-test="tag-list"
  >
    <template #rows>
      <tr
        v-for="(item, i) in tags"
        :key="i"
      >
        <td
          class="text-center"
          data-test="tag-name"
        >
          {{ item.name }}
        </td>
        <td class="text-center">
          <v-menu
            location="bottom"
            scrim
            eager
          >
            <template #activator="{ props }">
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
            <v-list
              class="bg-v-theme-surface"
              lines="two"
              density="compact"
            >
              <v-tooltip
                location="bottom"
                class="text-center"
                :disabled="canEditTag"
              >
                <template #activator="{ props }">
                  <div v-bind="props">
                    <TagEdit
                      :tag-name="item.name"
                      :has-authorization="canEditTag"
                      @update="refresh()"
                    />
                  </div>
                </template>
                <span> You don't have this kind of authorization. </span>
              </v-tooltip>

              <v-tooltip
                location="bottom"
                class="text-center"
                :disabled="canRemoveTag"
              >
                <template #activator="{ props }">
                  <div v-bind="props">
                    <TagRemove
                      :tag-name="item.name"
                      :has-authorization="canRemoveTag"
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
import hasPermission from "../../utils/permission";
import DataTable from "../Tables/DataTable.vue";
import TagRemove from "./TagRemove.vue";
import TagEdit from "./TagEdit.vue";
import handleError from "@/utils/handleError";
import useTagsStore from "@/store/modules/tags";
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

const tagsStore = useTagsStore();
const snackbar = useSnackbar();
const loading = ref(false);
const itemsPerPage = ref(10);
const page = ref<number>(1);
const tags = computed(() => tagsStore.list);
const tenant = computed(() => localStorage.getItem("tenant"));
const numberTags = computed<number>(
  () => tagsStore.getNumberTags,
);

const getTags = async (): Promise<void> => {
  if (!tenant.value) return;
  loading.value = true;
  try {
    await tagsStore.fetch({
      tenant: tenant.value,
      filter: tagsStore.getFilter || "",
      perPage: itemsPerPage.value,
      page: page.value,
    });
  } catch (error) {
    snackbar.showError("Failed to load tags.");
    handleError(error);
  } finally {
    loading.value = false;
  }
};

const refresh = async () => {
  await getTags();
};

watch([page, itemsPerPage], async () => { await getTags(); });

const canEditTag = hasPermission("tag:edit");

const canRemoveTag = hasPermission("tag:remove");

onMounted(async () => {
  await getTags();
});

defineExpose({ refresh });
</script>
