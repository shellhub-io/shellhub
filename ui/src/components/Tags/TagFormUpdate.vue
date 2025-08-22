<template>
  <v-list-item
    v-bind="$attrs"
    @click="open"
    :disabled="!hasAuthorization"
    data-test="open-tags-btn"
  >
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon> mdi-tag </v-icon>
      </div>

      <v-list-item-title data-test="has-tags-verification">
        {{ hasTags ? "Edit tags" : "Add Tags" }}
      </v-list-item-title>
    </div>
  </v-list-item>

  <BaseDialog v-model="showDialog">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-4 bg-primary" data-test="title">
        {{ hasTags ? "Edit tags" : "Add Tags" }}
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-5 w-100">
        <v-combobox
          id="targetInput"
          full-width
          v-model="inputTags"
          :error-messages="tagsError"
          label="Tag"
          hint="Maximum of 3 tags"
          multiple
          clearable
          chips
          variant="outlined"
          data-test="deviceTag-combobox"
          closable-chips
          :delimiters="[',', ' ']"
        />
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn
          variant="text"
          data-test="close-btn"
          @click="close()"
          class="mr-2"
        >
          Close
        </v-btn>

        <v-btn variant="text" data-test="save-btn" color="primary" @click="save()">
          Save
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import axios, { AxiosError } from "axios";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
import useDevicesStore from "@/store/modules/devices";
import useTagsStore from "@/store/modules/tags";

const props = defineProps<{
  deviceUid: string;
  tagsList: string[];
  hasAuthorization: boolean;
}>();

const emit = defineEmits(["update"]);
const snackbar = useSnackbar();
const devicesStore = useDevicesStore();
const tagsStore = useTagsStore();
const showDialog = ref(false);
const hasTags = computed(() => props.tagsList.length > 0);
const inputTags = ref<string[]>([]);
const tagsError = ref("");

const tagsHasLessThan3Characters = computed(() => inputTags.value.some((tag) => tag.length < 3));

watch(inputTags, () => {
  if (inputTags.value.length > 3) {
    tagsError.value = "Maximum of 3 tags";
  } else if (tagsHasLessThan3Characters.value) {
    tagsError.value = "The minimum length is 3 characters";
  } else {
    tagsError.value = "";
  }
});

const open = () => {
  inputTags.value.splice(0, inputTags.value.length, ...props.tagsList);
  showDialog.value = true;
};

const save = async () => {
  if (tagsError.value) return;
  try {
    tagsError.value = "";

    await devicesStore.updateDeviceTags({
      uid: props.deviceUid,
      tags: { tags: inputTags.value },
    });

    tagsStore.fetchTags();
    showDialog.value = false;
    snackbar.showSuccess("Tags updated successfully.");

    emit("update");
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      switch (axiosError.response?.status) {
        // when the name the format is invalid.
        case 400: {
          tagsError.value = "The format is invalid. Min 3, Max 255 characters!";
          break;
        }
        // when the user is not authorized.
        case 403: {
          snackbar.showError("You are not authorized to update this tag.");
          break;
        }
        // When the array tag size reached the max capacity.
        case 406: {
          tagsError.value = "The maximum capacity has reached.";
          break;
        }
        default: {
          snackbar.showError("Failed to update tags.");
          handleError(axiosError);
        }
      }
    } else {
      snackbar.showError("Failed to update tags.");
      handleError(error);
    }
  }
};

const close = () => {
  showDialog.value = false;
};
</script>
