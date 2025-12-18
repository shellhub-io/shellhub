<template>
  <PageHeader
    icon="mdi-bullhorn"
    title="Announcements"
    overline="Platform Messaging"
    description="Share important system broadcasts with every namespace administrator."
    icon-color="primary"
    data-test="announcement-title"
  >
    <template #actions>
      <v-btn
        :tabindex="0"
        data-test="new-announcement-btn"
        color="primary"
        append-icon="mdi-plus"
        text="New"
        @click="goToNewAnnouncement"
      />
    </template>
  </PageHeader>
  <AnnouncementList />
</template>

<script setup lang="ts">
import { useRouter } from "vue-router";
import PageHeader from "@/components/PageHeader.vue";
import AnnouncementList from "../components/Announcement/AnnouncementList.vue";

const router = useRouter();
const goToNewAnnouncement = async () => {
  await router.push({ name: "new-announcement" });
};
</script>
