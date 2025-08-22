<template>
  <v-list-item
    @click="open"
    v-bind="$attrs"
    :disabled="!hasAuthorization"
    data-test="open-tag-edit"
  >
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon> mdi-pencil </v-icon>
      </div>

      <v-list-item-title data-test="mdi-information-list-item">
        Edit
      </v-list-item-title>
    </div>
  </v-list-item>

  <BaseDialog v-model="showDialog">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-4 bg-primary"> Update Tag </v-card-title>
      <v-divider />

      <v-card-text class="mt-4 mb-0 pb-1">
        <v-text-field
          v-model="inputTags"
          label="Tag name"
          :error-messages="tagsError"
          required
          variant="underlined"
          data-test="tag-field"
        />
      </v-card-text>

      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" data-test="close-btn" @click="close()">
          Close
        </v-btn>

        <v-btn
          color="primary"
          variant="text"
          data-test="edit-btn"
          @click="edit()"
          :disabled="!!tagsError"
        >
          Edit
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
import useTagsStore from "@/store/modules/tags";

const props = defineProps<{
  tag: string;
  hasAuthorization: boolean;
}>();

const emit = defineEmits(["update"]);
const tagsStore = useTagsStore();
const snackbar = useSnackbar();
const showDialog = ref(false);

const inputTags = ref<string>("");
const tagsError = ref("");
const tagsHasLessThan3Characters = computed(() => inputTags.value.length < 3);

watch(inputTags, () => {
  if (inputTags.value.length > 255) {
    tagsError.value = "The maximum length is 255 characters";
  } else if (tagsHasLessThan3Characters.value) {
    tagsError.value = "The minimum length is 3 characters";
  } else {
    tagsError.value = "";
  }
});

const open = () => {
  inputTags.value = props.tag;
  showDialog.value = true;
};

const close = () => {
  showDialog.value = false;
};

const update = () => {
  emit("update");
  close();
};

const edit = async () => {
  if (!tagsError.value) {
    try {
      await tagsStore.updateTag({
        oldTag: props.tag,
        newTag: inputTags.value,
      });

      update();
      snackbar.showSuccess("Tag updated successfully.");
    } catch (error: unknown) {
      snackbar.showError("Failed to update tag.");
      handleError(error);
    }
  }
};

defineExpose({ inputTags });
</script>
