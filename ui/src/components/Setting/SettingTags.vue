<template>
  <TagCreate
    v-model="createDialog"
    @update="refreshTagList()"
  />
  <v-container fluid>
    <v-card
      variant="flat"
      class="bg-transparent"
      data-test="tags-settings-card"
    >
      <v-row cols="12">
        <v-col cols="3">
          <v-card-item class="pa-0 ma-0 mb-2">
            <v-list-item data-test="profile-header">
              <template #title>
                <h1>Tags</h1>
              </template>
              <template #subtitle>
                <span data-test="profile-subtitle">Manage your device and connector tags</span>
              </template>
            </v-list-item>
          </v-card-item>
        </v-col>
        <v-col cols="6">
          <v-text-field
            v-model.trim="filter"
            label="Search by Tag Name"
            variant="outlined"
            color="primary"
            single-line
            hide-details
            prepend-inner-icon="mdi-magnify"
            density="compact"
            data-test="search-text"
            @keyup="searchTags"
          />
        </v-col>
        <v-col
          cols="3"
          class="d-flex justify-end"
        >
          <v-btn
            color="primary"
            variant="elevated"
            data-test="tag-create-button"
            @click="openCreate"
          >
            Create Tag
          </v-btn>
        </v-col>
      </v-row>
      <TagList ref="tagListRef" />
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import TagList from "../Tags/TagList.vue";
import TagCreate from "../Tags/TagCreate.vue";
import useTagsStore from "@/store/modules/tags";
import useSnackbar from "@/helpers/snackbar";

const tagsStore = useTagsStore();
const snackbar = useSnackbar();
const tagListRef = ref<InstanceType<typeof TagList> | null>(null);
const createDialog = ref(false);
const filter = ref("");
const tenant = computed(() => localStorage.getItem("tenant") || "");

const searchTags = async () => {
  let encodedFilter = "";

  if (filter.value) {
    const filterToEncodeBase64 = [
      {
        type: "property",
        params: { name: "name", operator: "contains", value: filter.value },
      },
    ];
    encodedFilter = btoa(JSON.stringify(filterToEncodeBase64));
  }

  try {
    await tagsStore.search({
      tenant: tenant.value,
      filter: encodedFilter,
    });
  } catch {
    snackbar.showError("Failed to search tags.");
  }
};

const openCreate = () => {
  createDialog.value = true;
};

const refreshTagList = () => {
  tagListRef.value?.refresh();
};
</script>
