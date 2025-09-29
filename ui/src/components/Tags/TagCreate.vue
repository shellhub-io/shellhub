<template>
  <FormDialog
    v-model="showDialog"
    @close="close"
    @cancel="close"
    @confirm="create"
    title="Create Tag"
    icon="mdi-tag"
    confirm-text="Create"
    cancel-text="Close"
    :confirm-disabled="confirmDisabled"
    confirm-data-test="create-btn"
    cancel-data-test="close-btn"
    data-test="tag-create-dialog"
  >
    <div class="px-6 pt-4">
      <v-text-field
        v-model="inputTags"
        label="Tag name"
        :error-messages="tagsError"
        required
        data-test="tag-field"
      />
    </div>
  </FormDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useTagsStore from "@/store/modules/tags";
import FormDialog from "../FormDialog.vue";

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

const confirmDisabled = computed(() => !inputTags.value || !!tagsError.value);

const close = () => {
  inputTags.value = "";
  showDialog.value = false;
};

const update = () => {
  emit("update");
  close();
};

const create = async () => {
  if (tagsError.value) return;

  try {
    await tagsStore.createTag({
      tenant: tenant.value || "",
      name: inputTags.value,
    });

    snackbar.showSuccess("Successfully created tag");
    update();
  } catch (error: unknown) {
    snackbar.showError("Failed to create tag.");
    handleError(error);
  }
};

defineExpose({ inputTags, showDialog });
</script>
