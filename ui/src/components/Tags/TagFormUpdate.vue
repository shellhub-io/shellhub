<template>
  <v-list-item
    v-bind="$attrs"
    @click="open"
    :disabled="notHasAuthorization"
    data-test="open-tags-btn"
  >
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon> mdi-tag </v-icon>
      </div>
      <v-list-item-title data-test="hastags-verification">
        {{ hasTags ? "Edit tags" : "Add Tags" }}
      </v-list-item-title>
    </div>
  </v-list-item>

  <v-dialog v-model="showDialog" @click:outside="close()" min-width="280" max-width="450">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-4 bg-primary" data-test="title">
        {{ hasTags ? "Edit tags" : "Add Tags" }}
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-5 w-100">
        <v-autocomplete
          v-model="selectedTags"
          :items="tags"
          item-title="name"
          item-value="name"
          label="Tag"
          multiple
          variant="outlined"
          data-test="deviceTag-autocomplete"
          @update:search="searchTags"
        >
          <template v-slot:item="{ item }">
            <v-list-item
              v-bind="props"
              :key="item.value"
              @click="updateTags(item.value)"
              :active="selectedTags.includes(item.value)"
              active-color="primary"
              data-test="tag-item"
            >
              <v-list-item-action>
                <v-checkbox
                  :model-value="selectedTags.includes(item.value)"
                  color="primary"
                  hide-details
                />
                <v-list-item-title><v-chip>{{ item.value }} </v-chip></v-list-item-title>
              </v-list-item-action>
            </v-list-item>

            <v-divider />
          </template>

          <template v-slot:prepend-item>
            <div class="d-flex justify-center">
              <v-btn
                v-if="validNewTag"
                @click="createTag"
                color="primary"
                variant="text"
                data-test="create-new-tag-btn"
              >
                Create New Tag
              </v-btn>

            </div>
          </template>

          <template v-slot:selection="{ item }">
            <v-chip
              :key="item.value"
              closable
              @click:close="removeTag(item.value)"
              data-test="selected-tags"
            >
              {{ item.value }}
            </v-chip>
          </template>

          <template v-slot:append-item>
            <div class="d-flex justify-center" v-if="hasMore">
              <v-btn
                @click="loadTags"
                variant="text"
                color="primary"
                data-test="load-more-tags-btn"
                :disabled="!hasMore"
              >
                Load More
              </v-btn>
            </div>
          </template>
        </v-autocomplete>

      </v-card-text>

      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" data-test="close-btn" @click="close">Close</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import axios, { AxiosError } from "axios";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";

interface Tag {
  name: string;
}

const props = defineProps<{
  deviceUid: string;
  tagsList: Tag[];
  notHasAuthorization?: boolean;
}>();

const store = useStore();
const snackbar = useSnackbar();
const tenant = computed(() => localStorage.getItem("tenant"));
const showDialog = ref(false);
const emit = defineEmits(["update"]);
const selectedTags = ref<Array<string>>(
  props.tagsList.map((tag) => (typeof tag === "object" && "name" in tag ? tag.name : "")),
);
const fetchedTags = ref<Array<string>>([]);
const tags = computed(() => fetchedTags.value);
const page = ref(1);
const perPage = ref(10);
const hasMore = ref(true);
const filter = ref("");
const tagsError = ref("");

const hasTags = computed(() => selectedTags.value.length > 0);

const close = () => {
  selectedTags.value = [];
  fetchedTags.value = [];
  showDialog.value = false;
};

const encodeFilter = (search: string) => {
  if (!search) return "";
  const filterToEncodeBase64 = [
    {
      type: "property",
      params: { name: "name", operator: "contains", value: search },
    },
  ];
  return btoa(JSON.stringify(filterToEncodeBase64));
};

watch(selectedTags, (newValue) => {
  if (selectedTags.value === newValue) {
    emit("update");
  }
});

const validNewTag = computed(() => (
  filter.value.length >= 3
    && filter.value.length <= 255
    && !tags.value.includes(filter.value)
    && !selectedTags.value.includes(filter.value)
));

const loadTags = async () => {
  if (!hasMore.value) return;

  try {
    const encodedFilter = encodeFilter(filter.value);

    await store.dispatch("tags/autocomplete", {
      tenant: tenant.value,
      filter: encodedFilter,
      page: page.value,
      perPage: perPage.value,
    });

    const newTags = store.getters["tags/list"];

    if (newTags.length < perPage.value) hasMore.value = false;
    else page.value += 1; // Increment page to load next batch
    fetchedTags.value = [...fetchedTags.value, ...newTags]; // Append instead of replacing
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      switch (axiosError.response?.status) {
        // when the name the format is invalid.
        case 400: {
          tagsError.value = "The format is invalid. Min 3, Max 255 characters!";
          break;
        }
        // when the user is not authorized.
        case 403: {
          snackbar.showError("You are not authorized to update this tag.");
          break;
        }
        // When the array tag size reached the max capacity.
        case 406: {
          tagsError.value = "The maximum capacity has reached.";
          break;
        }
        default: {
          snackbar.showError("Failed to update tags.");
          handleError(axiosError);
        }
      }
    } else {
      snackbar.showError("Failed to update tags.");
      handleError(error);
    }
  }
};

const open = async () => {
  showDialog.value = true;
  page.value = 1;
  perPage.value = 10;
  hasMore.value = true;
  selectedTags.value = props.tagsList.map((tag) => typeof tag === "object" && "name" in tag ? tag.name : "");
  await loadTags();
};

const searchTags = (search) => {
  filter.value = search;

  const encodedFilter = encodeFilter(search);

  try {
    store.dispatch("tags/search", {
      tenant: tenant.value,
      filter: encodedFilter,
    });
  } catch {
    store.dispatch("snackbar/showSnackbarErrorDefault");
  }
};

const updateTags = async (newTag: string) => {
  const addedTag = !selectedTags.value.includes(newTag);

  if (addedTag) {
    await store.dispatch("tags/pushTagToDevice", {
      tenant: tenant.value,
      uid: props.deviceUid,
      name: newTag,
    });
    selectedTags.value.push(newTag);
  } else {
    await store.dispatch("tags/removeTagFromDevice", {
      tenant: tenant.value,
      uid: props.deviceUid,
      name: newTag,
    });
    const index = selectedTags.value.indexOf(newTag);
    selectedTags.value.splice(index, 1);
  }
};

const createTag = async () => {
  if (!validNewTag.value) return;

  try {
    await store.dispatch("tags/createTag", {
      tenant: tenant.value,
      name: filter.value,
    });

    await store.dispatch("tags/pushTagToDevice", {
      tenant: tenant.value,
      uid: props.deviceUid,
      name: filter.value,
    });

    selectedTags.value.push(filter.value);

    fetchedTags.value = [...new Set([filter.value, ...fetchedTags.value])];

    filter.value = "";
    tagsError.value = "";
  } catch (error) {
    console.error("Error creating tag:", error);
  }
};

const removeTag = async (tag) => {
  selectedTags.value = selectedTags.value.filter((t) => t !== tag);

  await store.dispatch("tags/removeTagFromDevice", {
    tenant: tenant.value,
    uid: props.deviceUid,
    name: tag,
  });
};

defineExpose({ updateTags, loadTags, createTag, removeTag, selectedTags, fetchedTags });

</script>
