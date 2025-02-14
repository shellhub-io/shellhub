<template>
  <v-tooltip bottom anchor="bottom">
    <template v-slot:activator="{ props }">
      <v-icon
        @click="dialog = !dialog"
        tag="a"
        dark
        v-bind="props"
        tabindex="0"
        aria-label="Dialog edit announcement"
        @keypress.enter="dialog = !dialog"
      >mdi-pencil
      </v-icon>
    </template>
    <span>Edit</span>
  </v-tooltip>

  <v-dialog
    v-model="dialog"
    max-width="70vw"
    persistent
    :retain-focus="false"
    :eager="true"
    transition="dialog-bottom-transition"
    z-index="1000"
  >
    <v-card>
      <v-card-title class="text-h5 pb-2"> Edit Announcement </v-card-title>
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
                    The announcement cannot be empty !
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
            @click="dialog = false"
            aria-label="Cancel"
          >Cancel
          </v-btn>
          <v-btn
            text
            type="submit"
            aria-label="Edit"
          >Submit
          </v-btn>
        </v-card-actions>
      </form>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { computed, defineComponent, ref, watch } from "vue";
import Editor from "@tinymce/tinymce-vue";
import { useField } from "vee-validate";
import MarkdownIt from "markdown-it";
import TurndownService from "turndown";
import * as yup from "yup";
import { useStore } from "../../store";
import { INotificationsError, INotificationsSuccess } from "../../interfaces/INotifications";
import { envVariables } from "../../envVariables";

export default defineComponent({
  name: "AnnouncementEdit",
  props: {
    announcement: {
      type: Object,
      required: true,
    },
  },
  emits: ["update"],
  setup(props, ctx) {
    const store = useStore();
    const dialog = ref(false);
    const md = new MarkdownIt();
    const turndownService = new TurndownService();
    const tinyMceKey = computed(() => envVariables.tinyMceKey);
    const tinyMceKeyIsEmpty = computed(() => tinyMceKey.value === "");
    const announcement = computed(() => store.getters["announcement/announcement"]);
    const contentInHtml = ref("");
    const contentError = ref(false);

    const {
      value: title,
      errorMessage: titleError,
    } = useField<string | undefined>("title", yup.string().required(), {
      initialValue: props.announcement.title,
    });

    const getAnnouncement = async () => {
      await store.dispatch("announcement/getAnnouncement", props.announcement.uuid);
      contentInHtml.value = md.render(announcement.value.content);
    };

    watch(dialog, async (val) => {
      if (val) {
        await getAnnouncement();
      }
    });

    watch(contentInHtml, () => {
      if (contentInHtml.value) {
        contentError.value = false;
      }
    });

    const onSubmit = async () => {
      if (titleError.value || !contentInHtml.value) {
        contentError.value = true;
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.announcementEdit,
        );
        return;
      }

      try {
        const contentInMarkdown = turndownService.turndown(contentInHtml.value);
        await store.dispatch("announcement/updateAnnouncement", {
          uuid: announcement.value.uuid,
          announcement: {
            title: title.value,
            content: contentInMarkdown,
          },
        });
        store.dispatch(
          "snackbar/showSnackbarSuccessAction",
          INotificationsSuccess.announcementEdit,
        );
        dialog.value = false;
        ctx.emit("update");
      } catch (error) {
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.announcementEdit,
        );
      }
    };

    return {
      dialog,
      tinyMceKey,
      title,
      titleError,
      contentInHtml,
      contentError,
      tinyMceKeyIsEmpty,
      onSubmit,
    };
  },
  components: {
    Editor,
  },
});
</script>

<style lang="scss">
.tox .tox-notification--warn, .tox .tox-notification--warning {
  display: none !important;
}
</style>
