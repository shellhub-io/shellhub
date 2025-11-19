<template>
  <TagCreate
    v-model="createDialog"
    @update="refreshTagList()"
  />
  <v-container
    fluid
    class="mx-0 px-0"
    max-width="60rem"
  >
    <div
      v-if="loading"
      class="d-flex justify-center mt-4"
    >
      <v-progress-circular
        indeterminate
        color="primary"
      />
    </div>

    <v-card
      v-else
      variant="flat"
      class="bg-transparent"
      data-test="tags-settings-card"
    >
      <v-card-item>
        <v-row cols="12">
          <v-col
            cols="3"
            class="pt-0"
          >
            <v-list-item
              class="pa-0 ma-0 mb-2"
              data-test="profile-header"
            >
              <template #title>
                <h1>Tags</h1>
              </template>
              <template #subtitle>
                <span data-test="profile-subtitle">Manage your device and connector tags</span>
              </template>
            </v-list-item>
          </v-col>

          <v-col cols="6">
            <v-text-field
              v-if="hasTags"
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
              v-if="hasTags"
              color="primary"
              variant="elevated"
              data-test="tag-create-button"
              @click="openCreate"
            >
              Create Tag
            </v-btn>
          </v-col>
        </v-row>
      </v-card-item>

      <TagList
        v-if="hasTags"
        ref="tagListRef"
        class="mx-4"
      />

      <NoItemsMessage
        v-else
        item="Tags"
        icon="mdi-tag-multiple"
        data-test="no-items-message-component"
      >
        <template #content>
          <p>
            ShellHub allows you to organize your resources using Tags.
          </p>
          <p>
            You can assign tags to Devices, Public Keys, and Firewall Rules
            to filter and group them effectively.
          </p>
        </template>
        <template #action>
          <v-btn
            color="primary"
            variant="elevated"
            @click="openCreate"
          >
            Create Tag
          </v-btn>
        </template>
      </NoItemsMessage>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from "vue";
import TagList from "../Tags/TagList.vue";
import TagCreate from "../Tags/TagCreate.vue";
import NoItemsMessage from "@/components/NoItemsMessage.vue";
import useTagsStore from "@/store/modules/tags";
import useSnackbar from "@/helpers/snackbar";
import handleError from "@/utils/handleError";

const tagsStore = useTagsStore();
const snackbar = useSnackbar();
const tagListRef = ref<InstanceType<typeof TagList> | null>(null);
const createDialog = ref(false);
const filter = ref("");
const loading = ref(true);
const tenant = computed(() => localStorage.getItem("tenant") || "");

const hasTags = computed(() => tagsStore.getNumberTags > 0);

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

const refreshTagList = async () => {
  if (!hasTags.value) {
    await fetchInitialTags();
  } else {
    tagListRef.value?.refresh();
  }
};

const fetchInitialTags = async () => {
  try {
    loading.value = true;
    await tagsStore.fetch({
      tenant: tenant.value,
      page: 1,
      perPage: 10,
      filter: "",
    });
  } catch (error) {
    snackbar.showError("Failed to load tags.");
    handleError(error);
  } finally {
    loading.value = false;
  }
};

onMounted(async () => {
  await fetchInitialTags();
});
</script>
