<template>
  <div class="d-flex flex-column justify-space-between align-center flex-md-row mb-2">
    <h1 data-test="announcement-title">Announcements</h1>
    <v-spacer />
    <v-btn
      :tabindex="0"
      @click="newAnnouncement"
      data-test="new-announcement-btn"
      color="primary"
      append-icon="mdi-plus"
      text="New"
    />
  </div>
  <AnnouncementList />
</template>

<script setup lang="ts">
import { useRouter } from "vue-router";
import AnnouncementList from "../components/Announcement/AnnouncementList.vue";

const router = useRouter();
const newAnnouncement = () => {
  router.push({ name: "new-announcement" });
};
</script>
