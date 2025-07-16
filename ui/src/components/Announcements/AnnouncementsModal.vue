<template>
  <BaseDialog
    v-model="showDialog"
    scrollable
    persistent
  >
    <v-card class="bg-grey-darken-4 bg-v-theme-surface">
      <v-card-title class="bg-primary" data-test="announcement-title">
        {{ announcement.title }}
      </v-card-title>
      <v-divider />

      <v-card-text style="max-height: 70vh">
        <div class="content-announcement" v-html="markdownContent" data-test="announcement-title" />

        <div class="text-right">
          <span
            class="text-caption text-medium-emphasis"
            data-test="announcement-date"
          >
            Posted in {{ date }}
          </span>
        </div>
      </v-card-text>
      <v-divider />

      <v-card-actions class="pr-5">
        <v-spacer />
        <v-btn
          @click="close"
          color="primary"
          variant="elevated"
          tabindex="0"
          data-test="announcement-close"
        >
          Dismiss
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from "vue";
import MarkdownIt from "markdown-it";
import moment from "moment";
import BaseDialog from "../BaseDialog.vue";
import { IAnnouncement } from "@/interfaces/IAnnouncement";

const props = defineProps<{ announcement: IAnnouncement }>();

const showDialog = defineModel<boolean>({ required: true });
const md = new MarkdownIt();
const date = computed(() => moment(props.announcement.date).format("LL"));
const markdownContent = computed(() => md.render(props.announcement.content));

const close = () => {
  localStorage.setItem("announcement", btoa(JSON.stringify(props.announcement)));
  showDialog.value = false;
};
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
