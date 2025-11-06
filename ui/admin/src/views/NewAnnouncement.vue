<template>
  <h1>Create new Announcement</h1>

  <v-card class="pa-4 mt-2 bg-background border">
    <v-card-item>
      <v-card-title>Title</v-card-title>

      <v-text-field
        v-model="title"
        :error-messages="titleError"
        variant="underlined"
        placeholder="Enter announcement title"
        data-test="announcement-title"
      />
    </v-card-item>

    <v-card-item class="mt-n4">
      <v-card-title>Content</v-card-title>

      <Editor
        v-model="announcement"
        :api-key="tinyMceKey"
        output-format="html"
        :init="{
          plugins: 'lists link image code help wordcount',
          menubar: 'file edit insert view tools help',
        }"
        toolbar="undo redo | blocks | fontsize image | bold italic link blockquote |
          \ bullist numlist | removeformat | help"
        data-test="announcement-content"
      />

      <v-alert
        v-if="announcementError"
        type="error"
        class="mt-2"
        data-test="announcement-error"
      >
        The announcement cannot be empty!
      </v-alert>

      <v-alert
        v-if="tinyMceKeyIsEmpty"
        type="warning"
        class="mt-2"
        data-test="announcement-key-warning"
      >
        It's recommended to set the TinyMCE key in the .env file.
      </v-alert>
    </v-card-item>

    <v-card-actions class="pa-4">
      <v-spacer />
      <v-btn
        color="primary"
        variant="elevated"
        tabindex="0"
        data-test="announcement-btn-post"
        text="Post"
        @click="postAnnouncement"
      />
    </v-card-actions>
  </v-card>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import Editor from "@tinymce/tinymce-vue";
import TurndownService from "turndown";
import { useField } from "vee-validate";
import * as yup from "yup";
import { useRouter } from "vue-router";
import useAnnouncementStore from "@admin/store/modules/announcement";
import useSnackbar from "@/helpers/snackbar";
import { envVariables } from "../envVariables";
import handleError from "@/utils/handleError";

const router = useRouter();
const snackbar = useSnackbar();
const announcementStore = useAnnouncementStore();
const { value: title, errorMessage: titleError, setErrors: setTitleError } = useField<
      string | undefined
>("title", yup.string().required(), {
  initialValue: "",
});
const tinyMceKey = computed(() => envVariables.tinyMceKey);
const tinyMceKeyIsEmpty = computed(() => tinyMceKey.value === "");
const announcement = ref("");
const announcementError = ref(false);
// eslint-disable-next-line @typescript-eslint/no-unsafe-call
const { turndown } = new TurndownService() as { turndown: (input: string) => string };

watch(announcement, (val) => {
  if (val) announcementError.value = false;
});

const postAnnouncement = async () => {
  if (!title.value) {
    setTitleError("Title cannot be empty!");
    return;
  }

  if (titleError.value || !announcement.value) {
    announcementError.value = true;
    snackbar.showError("Failed to create announcement.");
    return;
  }

  try {
    const contentInHtml = turndown(announcement.value);
    await announcementStore.createAnnouncement({
      title: title.value,
      content: contentInHtml,
    });
    snackbar.showSuccess("Successfully created announcement.");
    await router.push({ name: "announcements" });
  } catch (error) {
    handleError(error);
    snackbar.showError("Failed to create announcement.");
  }
};

defineExpose({ title, titleError });
</script>

<style lang="scss">
.tox .tox-notification--warn, .tox .tox-notification--warning {
  display: none !important;
}
</style>
