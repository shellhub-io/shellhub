<template>
  <TagCreate
    v-model="showCreateTagDialog"
    @update="refreshTagList()"
  />

  <PageHeader
    icon="mdi-tag-multiple"
    title="Tags"
    overline="Organization"
    :description="tagsDescription"
    icon-color="primary"
    data-test="tags-settings-card"
  >
    <template #actions>
      <v-text-field
        v-if="showTags"
        v-model.trim="filter"
        class="w-sm-50"
        label="Search by Tag Name"
        variant="outlined"
        color="primary"
        single-line
        hide-details
        prepend-inner-icon="mdi-magnify"
        density="compact"
        data-test="search-text"
      />
      <v-btn
        v-if="showTags"
        color="primary"
        variant="elevated"
        data-test="tag-create-button"
        text="Create Tag"
        @click="openCreateTagDialog"
      />
    </template>
  </PageHeader>

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
import PageHeader from "@/components/PageHeader.vue";
import useTagsStore from "@/store/modules/tags";
const tagsDescription = "Organize and categorize your resources with tags. "
  + "Apply tags to devices, public keys, and firewall rules for easier filtering and management.";

const tagsStore = useTagsStore();
const tagListRef = ref<InstanceType<typeof TagList> | null>(null);
const showCreateTagDialog = ref(false);
const filter = ref("");
const showTags = computed(() => tagsStore.showTags);

const openCreateTagDialog = () => { showCreateTagDialog.value = true; };

const refreshTagList = async () => { await tagListRef.value?.getTags(); };
</script>
