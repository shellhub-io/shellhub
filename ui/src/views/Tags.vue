<template>
  <TagCreate
    v-model="showCreateTagDialog"
    @update="refreshTagList()"
  />

  <div
    class="d-flex flex-column justify-space-between align-center flex-sm-row w-100 ga-1 mb-4"
    data-test="tags-settings-card"
  >
    <h1>Tags</h1>
    <v-spacer />
    <v-text-field
      v-if="showTags"
      v-model.trim="filter"
      class="w-100 w-sm-50"
      label="Search by Tag Name"
      color="primary"
      single-line
      hide-details
      prepend-inner-icon="mdi-magnify"
      density="compact"
      data-test="search-text"
    />
    <v-spacer />
    <v-btn
      v-if="showTags"
      color="primary"
      variant="elevated"
      data-test="tag-create-button"
      text="Create Tag"
      @click="openCreateTagDialog"
    />
  </div>

  <TagList
    v-if="showTags"
    ref="tagListRef"
    :filter
  />

  <NoItemsMessage
    v-else
    item="Tags"
    icon="mdi-tag-multiple"
    data-test="no-items-message-component"
  >
    <template #content>
      <p>ShellHub allows you to organize your resources using Tags.</p>
      <p>
        You can assign tags to Devices, Public Keys, and Firewall Rules
        to filter and group them effectively.
      </p>
    </template>
    <template #action>
      <v-btn
        color="primary"
        variant="elevated"
        text="Create Tag"
        @click="openCreateTagDialog"
      />
    </template>
  </NoItemsMessage>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import TagList from "@/components/Tags/TagList.vue";
import TagCreate from "@/components/Tags/TagCreate.vue";
import NoItemsMessage from "@/components/NoItemsMessage.vue";
import useTagsStore from "@/store/modules/tags";

const tagsStore = useTagsStore();
const tagListRef = ref<InstanceType<typeof TagList> | null>(null);
const showCreateTagDialog = ref(false);
const filter = ref("");
const showTags = computed(() => tagsStore.showTags);

const openCreateTagDialog = () => { showCreateTagDialog.value = true; };

const refreshTagList = async () => { await tagListRef.value?.getTags(); };
</script>
