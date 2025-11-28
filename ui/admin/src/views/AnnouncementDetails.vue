<template>
  <h1>Announcement Details</h1>
  <v-card
    v-if="announcement?.uuid"
    class="mt-2 border rounded bg-background"
  >
    <v-card-title class="pa-4 d-flex align-center justify-space-between bg-v-theme-surface">
      <h2 class="text-h6 ml-2">{{ announcement.title }}</h2>

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
            data-test="announcement-actions-menu-btn"
          />
        </template>

        <v-list
          class="bg-v-theme-surface"
          lines="two"
          density="compact"
        >
          <AnnouncementEdit
            v-slot="{ openDialog }"
            :announcement-item="announcement"
            @update="loadAnnouncement"
          >
            <v-list-item
              data-test="announcement-edit-btn"
              @click="openDialog"
            >
              <div class="d-flex align-center">
                <v-icon
                  class="mr-2"
                  icon="mdi-pencil"
                />
                <v-list-item-title>Edit this announcement</v-list-item-title>
              </div>
            </v-list-item>
          </AnnouncementEdit>

          <AnnouncementDelete
            v-slot="{ openDialog }"
            :uuid="announcement.uuid"
            redirect
            @update="handleDelete"
          >
            <v-list-item
              data-test="announcement-delete-btn"
              @click="openDialog"
            >
              <div class="d-flex align-center">
                <v-icon
                  class="mr-2"
                  icon="mdi-delete"
                />
                <v-list-item-title>Delete this announcement</v-list-item-title>
              </div>
            </v-list-item>
          </AnnouncementDelete>

          <div class="px-2 py-1" />
        </v-list>
      </v-menu>
    </v-card-title>

    <v-divider />

    <v-card-text class="pa-4 pt-0">
      <v-row class="py-3">
        <v-col
          cols="12"
          md="6"
          class="my-0 py-0"
        >
          <div data-test="announcement-uuid-field">
            <h3 class="item-title">UUID:</h3>
            <p class="text-truncate">{{ announcement.uuid }}</p>
          </div>
        </v-col>

        <v-col
          cols="12"
          md="6"
          class="my-0 py-0"
        >
          <div data-test="announcement-date-field">
            <h3 class="item-title">Date:</h3>
            <p>{{ formatFullDateTime(announcement.date) }}</p>
          </div>
        </v-col>
      </v-row>
      <v-divider class="my-4" />
      <v-row class="px-3">
        <div data-test="announcement-content-field">
          <h3 class="item-title">Content:</h3>
          <div class="pa-2 content-announcement">
            <div v-html="contentToHtml" />
          </div>
        </div>
      </v-row>
    </v-card-text>
  </v-card>
  <v-card
    v-else
    class="mt-2 pa-4 bg-v-theme-surface"
  >
    <p class="text-center">Something is wrong, try again!</p>
  </v-card>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import MarkdownIt from "markdown-it";
import useAnnouncementStore from "@admin/store/modules/announcement";
import AnnouncementEdit from "@admin/components/Announcement/AnnouncementEdit.vue";
import AnnouncementDelete from "@admin/components/Announcement/AnnouncementDelete.vue";
import { formatFullDateTime } from "@/utils/date";
import useSnackbar from "@/helpers/snackbar";

const md = new MarkdownIt();
const announcementStore = useAnnouncementStore();
const route = useRoute();
const router = useRouter();
const snackbar = useSnackbar();

const announcementId = computed(() => route.params.uuid as string);
const announcement = computed(() => announcementStore.announcement);
const contentToHtml = ref("");

const loadAnnouncement = async () => {
  try {
    await announcementStore.fetchAnnouncement(announcementId.value);
    if (announcement.value) contentToHtml.value = md.render(announcement.value?.content || "");
  } catch {
    snackbar.showError("Failed to get announcement details.");
  }
};

const handleDelete = () => { void router.push({ name: "announcements" }); };

onMounted(loadAnnouncement);

defineExpose({ announcement, contentToHtml });
</script>

<style lang="scss" scoped>
:deep(.content-announcement) {
  p, span, div, h1, h2, h3, h4, h5, h6 {
    margin: .5rem;
    line-height: 1.4;
  }

  ul, ol {
    padding: .5rem 1.5rem;
  }

  img {
    max-width: 100%;
  }

  blockquote {
    display: block;
    margin-block-start: 1em;
    margin-block-end: 1em;
    margin-inline-start: 40px;
    margin-inline-end: 40px;
  }
}
</style>
