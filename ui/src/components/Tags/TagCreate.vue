<template>
  <BaseDialog v-model="showDialog" min-width="300" max-width="600">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-4 bg-primary"> Create Tag </v-card-title>
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
          data-test="create-btn"
          @click="create()"
        >
          Create
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useTagsStore from "@/store/modules/tags";
import BaseDialog from "../BaseDialog.vue";

const emit = defineEmits(["update"]);
const tagsStore = useTagsStore();
const snackbar = useSnackbar();
const showDialog = defineModel({ default: false });

const inputTags = ref<string>("");
const tagsError = ref("");
const tagsHasLessThan3Characters = computed(() => inputTags.value.length < 3);
const tenant = computed(() => localStorage.getItem("tenant"));

watch(inputTags, () => {
  if (inputTags.value.length > 255) {
    tagsError.value = "Maximum of 3 tags";
  } else if (tagsHasLessThan3Characters.value) {
    tagsError.value = "The minimum length is 3 characters";
  } else {
    tagsError.value = "";
  }
});

const close = () => {
  inputTags.value = "";
  showDialog.value = false;
};

const update = () => {
  emit("update");
  close();
};

const create = async () => {
  if (!tagsError.value) {
    try {
      await tagsStore.createTag({
        tenant: tenant.value || "",
        name: inputTags.value,
      });

      update();
      snackbar.showSuccess("Successfully created tag");
    } catch (error: unknown) {
      snackbar.showError("Failed to create tag.");
      handleError(error);
    }
  }
};

defineExpose({ inputTags, showDialog });
</script>
