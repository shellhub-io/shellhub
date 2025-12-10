<template>
  <DataTable
    v-model:items-per-page="itemsPerPage"
    v-model:page="page"
    :headers
    :items="tags"
    :loading
    :total-count="tagCount"
    :items-per-page-options="[10, 20, 50, 100]"
    data-test="tag-list"
  >
    <template #rows>
      <tr
        v-for="tag in tags"
        :key="tag.name"
      >
        <td
          class="text-center"
          data-test="tag-name"
        >
          {{ tag.name }}
        </td>
        <td
          class="text-center"
          data-test="tag-created-at"
        >
          {{ formatShortDateTime(tag.created_at) }}
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
                      :tag-name="tag.name"
                      :has-authorization="canEditTag"
                      @update="getTags"
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
                      :tag-name="tag.name"
                      :has-authorization="canRemoveTag"
                      @update="getTags"
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
import { formatShortDateTime } from "@/utils/date";

const props = defineProps<{ filter: string }>();

const headers = ref([
  {
    text: "Name",
    value: "name",
  },
  {
    text: "Created At",
    value: "created_at",
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
const page = ref(1);
const tags = computed(() => tagsStore.tags);
const tagCount = computed(() => tagsStore.tagCount);

const encodeFilter = () => {
  if (!props.filter) return undefined;
  const filterObject = [{
    type: "property",
    params: { name: "name", operator: "contains", value: props.filter },
  }];
  return Buffer.from(JSON.stringify(filterObject)).toString("base64");
};

const getTags = async () => {
  loading.value = true;
  try {
    await tagsStore.fetchTagList({
      perPage: itemsPerPage.value,
      page: page.value,
      filter: encodeFilter(),
    });
  } catch (error) {
    snackbar.showError("Failed to load tags.");
    handleError(error);
  } finally {
    loading.value = false;
  }
};

watch(() => props.filter, async () => {
  page.value = 1;
  await getTags();
}, { immediate: true });

watch([page, itemsPerPage], async () => { await getTags(); });

const canEditTag = hasPermission("tag:edit");

const canRemoveTag = hasPermission("tag:remove");

onMounted(async () => { await getTags(); });

defineExpose({ getTags });
</script>
