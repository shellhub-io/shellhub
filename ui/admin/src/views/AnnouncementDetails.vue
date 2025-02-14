<template>
  <div class="d-flex pa-0 align-center">
    <h1>Announcement Details</h1>
  </div>

  <v-card class="mt-2 pa-4">
    <v-card-text>
      <div>
        <div class="text-overline mt-3">
          <h3>Uuid:</h3>
        </div>
        <div :data-test="announcement.uuid">
          <p>{{ announcement.uuid }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Title:</h3>
        </div>
        <div :data-test="announcement.title">
          <p>{{ announcement.title }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Content:</h3>
        </div>
        <div class="pa-2">
          <div :data-test="announcement.content">
            <div class="content-announcement" v-html="contentToHtml" />
          </div>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Date:</h3>
        </div>
        <div :data-test="announcement.date">
          <p>{{ date }}</p>
        </div>
      </div>
    </v-card-text>
  </v-card>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import MarkdownIt from "markdown-it";
import moment from "moment";
import { useStore } from "../store";

export default defineComponent({
  name: "Announcement",
  setup() {
    const md = new MarkdownIt();
    const store = useStore();
    const route = useRoute();
    const announcementId = computed(() => route.params.uuid);
    const announcement = computed(
      () => store.getters["announcement/announcement"],
    );
    const contentToHtml = ref("");
    const date = ref("");

    onMounted(async () => {
      await store.dispatch(
        "announcement/getAnnouncement",
        announcementId.value,
      );

      if (announcement.value) {
        contentToHtml.value = md.render(announcement.value.content);
        date.value = moment(announcement.value.date).format("LL");
      }
    });

    return {
      announcement,
      contentToHtml,
      date,
    };
  },
});
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
