<template>
  <div
    class="d-flex flex-column justify-space-between align-center flex-md-row"
  >
    <h1 data-test="announcement-title">Announcements</h1>
    <v-spacer />
    <v-card class="mt-2">
      <v-btn
        :tabindex="0"
        @click="newAnnouncement"
        data-test="new-announcement-btn"
      >
        New
        <v-icon class="ml-1">mdi-plus</v-icon>
      </v-btn>
    </v-card>
  </div>

  <v-card class="mt-2" data-test="announcement-list">
    <AnnouncementList />
  </v-card>
</template>

<script setup lang="ts">
import { useRouter } from "vue-router";
import AnnouncementList from "../components/Announcement/AnnouncementList.vue";

const router = useRouter();
const newAnnouncement = () => {
  router.push({ name: "new-announcement" });
};
</script>
