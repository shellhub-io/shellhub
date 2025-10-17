<template>
  <v-tooltip bottom anchor="bottom">
    <template v-slot:activator="{ props }">
      <v-icon
        @click="open"
        tag="button"
        v-bind="props"
        tabindex="0"
        data-test="edit-button"
        icon="mdi-pencil"
      />
    </template>
    <span>Edit</span>
  </v-tooltip>

  <BaseDialog
    v-model="showDialog"
    persistent
    :retain-focus="false"
    :eager="true"
    transition="dialog-bottom-transition"
  >
    <v-card>
      <v-card-title class="text-h5 pb-2">Edit Announcement</v-card-title>
      <v-divider />
      <form @submit.prevent="onSubmit">
        <v-card-text>
          <v-container>
            <v-row>
              <v-col cols="12">
                <v-container>
                  <v-text-field
                    v-model="title"
                    label="Title"
                    required
                    :error-messages="titleError"
                    color="primary"
                    variant="underlined"
                  />

                  <Editor
                    :api-key="tinyMceKey"
                    v-model="contentInHtml"
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

                </v-container>
              </v-col>
            </v-row>
          </v-container>
        </v-card-text>
        <v-divider />
        <v-card-actions>
          <v-spacer />
          <v-btn
            text
            @click="showDialog = false"
            aria-label="Cancel"
          >Cancel
          </v-btn>
          <v-btn
            text
            type="submit"
            color="primary"
            aria-label="Edit"
          >Submit
          </v-btn>
        </v-card-actions>
      </form>
    </v-card>
  </BaseDialog>
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
import { envVariables } from "../../envVariables";
import handleError from "@/utils/handleError";
import BaseDialog from "@/components/BaseDialog.vue";

const props = defineProps<{ announcementItem: IAdminAnnouncementShort }>();

const emit = defineEmits(["update"]);
const announcementStore = useAnnouncementStore();
const snackbar = useSnackbar();
const showDialog = ref(false);
const md = new MarkdownIt();
const turndownService = new TurndownService();
const tinyMceKey = computed(() => envVariables.tinyMceKey);
const tinyMceKeyIsEmpty = computed(() => tinyMceKey.value === "");
const announcement = computed(() => announcementStore.announcement);
const contentInHtml = ref("");
const contentError = ref(false);

const {
  value: title,
  errorMessage: titleError,
} = useField<string | undefined>("title", yup.string().required(), {
  initialValue: props.announcementItem.title,
});

const getAnnouncement = async () => {
  await announcementStore.fetchAnnouncement(props.announcementItem.uuid);
  contentInHtml.value = md.render(announcement.value.content as string);
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
    const contentInMarkdown = turndownService.turndown(contentInHtml.value);
    await announcementStore.updateAnnouncement(announcement.value.uuid as string, { title: title.value ?? "", content: contentInMarkdown });
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
.tox .tox-notification--warn, .tox .tox-notification--warning {
  display: none !important;
}
</style>
