<template>
  <v-tooltip
    bottom
    anchor="bottom"
  >
    <template #activator="{ props }">
      <v-icon
        tag="button"
        v-bind="props"
        tabindex="0"
        data-test="edit-button"
        icon="mdi-pencil"
        @click="open"
      />
    </template>
    <span>Edit</span>
  </v-tooltip>

  <FormDialog
    v-model="showDialog"
    title="Edit Announcement"
    icon="mdi-bullhorn"
    icon-color="primary"
    confirm-text="Submit"
    cancel-text="Cancel"
    @close="showDialog = false"
    @confirm="onSubmit"
    @cancel="showDialog = false"
  >
    <v-card-text class="pa-6">
      <v-text-field
        v-model="title"
        label="Title"
        required
        :error-messages="titleError"
        color="primary"
      />
      <Editor
        v-model="contentInHtml"
        :api-key="tinyMceKey"
        :init="{
          plugins: 'lists link image code help wordcount',
          menubar: 'file edit insert view tools help',
        }"
        toolbar="undo redo | blocks | fontsize image | bold italic link blockquote |
                  \ bullist numlist | removeformat | help"
        data-test="announcement-content"
      />
      <v-alert
        v-if="contentError"
        type="error"
        class="mt-2"
        data-test="announcement-error"
        text="The announcement cannot be empty!"
      />
      <v-alert
        v-if="isTinyMceKeyEmpty"
        type="warning"
        class="mt-2"
        data-test="announcement-key-warning"
        text="It's recommended to set the TinyMCE key in the .env file."
      />
    </v-card-text>
  </FormDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import Editor from "@tinymce/tinymce-vue";
import { useField } from "vee-validate";
import MarkdownIt from "markdown-it";
import TurndownService from "turndown";
import * as yup from "yup";
import useAnnouncementStore from "@admin/store/modules/announcement";
import { IAdminAnnouncementShort } from "@admin/interfaces/IAnnouncement";
import useSnackbar from "@/helpers/snackbar";
import { envVariables } from "@/envVariables";
import handleError from "@/utils/handleError";
import FormDialog from "@/components/Dialogs/FormDialog.vue";

const props = defineProps<{ announcementItem: IAdminAnnouncementShort }>();

const emit = defineEmits(["update"]);
const announcementStore = useAnnouncementStore();
const snackbar = useSnackbar();
const showDialog = ref(false);
const md = new MarkdownIt();
// eslint-disable-next-line @typescript-eslint/no-unsafe-call
const { turndown } = new TurndownService() as {
  turndown: (input: string) => string;
};
const tinyMceKey = computed(() => envVariables.tinyMceKey);
const isTinyMceKeyEmpty = computed(() => tinyMceKey.value === "");
const announcement = computed(() => announcementStore.announcement);
const contentInHtml = ref("");
const contentError = ref(false);

const { value: title, errorMessage: titleError } = useField<string | undefined>(
  "title",
  yup.string().required(),
  {
    initialValue: props.announcementItem.title,
  },
);

const getAnnouncement = async () => {
  await announcementStore.fetchAnnouncement(props.announcementItem.uuid);
  contentInHtml.value = md.render(announcement.value.content);
};

const open = async () => {
  await getAnnouncement();
  showDialog.value = true;
};

watch(contentInHtml, () => {
  if (contentInHtml.value) {
    contentError.value = false;
  }
});

const onSubmit = async () => {
  if (titleError.value || !contentInHtml.value) {
    contentError.value = true;
    snackbar.showError("Please fill in all required fields.");
    return;
  }

  try {
    const contentInMarkdown = turndown(contentInHtml.value);
    await announcementStore.updateAnnouncement(announcement.value.uuid, {
      title: title.value ?? "",
      content: contentInMarkdown,
    });
    snackbar.showSuccess("Announcement updated successfully.");
    showDialog.value = false;
    emit("update");
  } catch (error) {
    handleError(error);
    snackbar.showError("Failed to update announcement.");
  }
};

defineExpose({ showDialog, announcement, contentInHtml, contentError, title });
</script>

<style lang="scss">
.tox .tox-notification--warn,
.tox .tox-notification--warning {
  display: none !important;
}
</style>
