<!-- eslint-disable vue/no-v-html -->
<template>
  <h1>Announcement Details</h1>
  <v-card class="mt-2 pa-4 bg-background border">
    <v-card-text>
      <div>
        <h3 class="text-overline">
          Uuid:
        </h3>
        <p :data-test="announcement.uuid">
          {{ announcement.uuid }}
        </p>
      </div>

      <div>
        <h3 class="text-overline mt-3">
          Title:
        </h3>
        <p :data-test="announcement.title">
          {{ announcement.title }}
        </p>
      </div>

      <div>
        <h3 class="text-overline mt-3">
          Content:
        </h3>
        <div
          class="pa-2"
          :data-test="announcement.content"
        >
          <div
            class="content-announcement"
            v-html="contentToHtml"
          />
        </div>
      </div>

      <div>
        <h3 class="text-overline mt-3">
          Date:
        </h3>
        <p :data-test="announcement.date">
          {{ date }}
        </p>
      </div>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import MarkdownIt from "markdown-it";
import moment from "moment";
import useAnnouncementStore from "@admin/store/modules/announcement";

const md = new MarkdownIt();
const announcementStore = useAnnouncementStore();
const route = useRoute();
const announcementId = computed(() => route.params.uuid);
const announcement = computed(() => announcementStore.announcement);
const contentToHtml = ref("");
const date = ref("");

onMounted(async () => {
  await announcementStore.fetchAnnouncement(announcementId.value as string);

  if (announcement.value) {
    contentToHtml.value = md.render(announcement.value?.content || "");
    date.value = moment(announcement.value.date).format("LL");
  }
});

defineExpose({ announcement, contentToHtml, date });
</script>

<style lang="scss">
.content-announcement {
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
