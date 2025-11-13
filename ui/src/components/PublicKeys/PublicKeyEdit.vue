<template>
  <div>
    <v-list-item
      v-bind="$attrs"
      :disabled="!hasAuthorization"
      data-test="public-key-edit-title-btn"
      @click="open()"
    >
      <div class="d-flex align-center">
        <div
          data-test="public-key-edit-icon"
          class="mr-2"
        >
          <v-icon icon="mdi-pencil" />
        </div>
        <v-list-item-title>Edit</v-list-item-title>
      </div>
    </v-list-item>

    <FormDialog
      v-model="showDialog"
      title="Edit Public Key"
      icon="mdi-key-outline"
      confirm-text="Save"
      cancel-text="Cancel"
      :confirm-disabled
      confirm-data-test="pk-edit-save-btn"
      cancel-data-test="pk-edit-cancel-btn"
      data-test="public-key-edit-dialog"
      @close="close"
      @cancel="close"
      @confirm="edit"
    >
      <div class="px-6 pt-4">
        <v-row class="mt-1 px-3">
          <v-text-field
            v-model="name"
            label="Key name"
            placeholder="Name used to identify the public key"
            :error-messages="nameError"
            required
            data-test="name-field"
            class="mb-5"
            hide-details="auto"
          />
        </v-row>

        <v-row class="mt-2 px-3">
          <v-select
            v-model="selectedUsernameOption"
            label="Device username access restriction"
            :items="usernameSelectOptions"
            data-test="username-restriction-field"
          />
        </v-row>

        <v-row class="mt-2 px-3">
          <v-text-field
            v-if="selectedUsernameOption === FormUsernameOptions.Username"
            v-model="username"
            label="Rule username"
            :error-messages="usernameError"
            data-test="rule-field"
          />
        </v-row>

        <v-row class="mt-4 px-3">
          <v-select
            v-model="selectedFilterOption"
            label="Device access restriction"
            :items="filterSelectOptions"
            data-test="filter-restriction-field"
          />
        </v-row>

        <v-row class="mt-1 px-3">
          <TagAutocompleteSelect
            v-if="selectedFilterOption === FormFilterOptions.Tags"
            v-model:selected-tags="selectedTags"
            v-model:tag-selector-error-message="tagSelectorErrorMessage"
          />

          <v-text-field
            v-else-if="selectedFilterOption === FormFilterOptions.Hostname"
            v-model="hostname"
            label="Hostname"
            :error-messages="hostnameError"
            data-test="hostname-field"
          />
        </v-row>

        <FileTextComponent
          v-model="publicKeyData"
          class="mt-4 mb-2"
          text-only
          textarea-label="Public key data"
          description-text="Public key data cannot be modified after creation."
          :disabled="true"
          data-test="data-field"
        />
      </div>
    </FormDialog>
  </div>
</template>

<script setup lang="ts">
import { useField } from "vee-validate";
import { ref, computed } from "vue";
import * as yup from "yup";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import { IPublicKey } from "@/interfaces/IPublicKey";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import { FormFilterOptions, FormUsernameOptions } from "@/interfaces/IFilter";
import usePublicKeysStore from "@/store/modules/public_keys";
import FileTextComponent from "@/components/Fields/FileTextComponent.vue";
import TagAutocompleteSelect from "@/components/Tags/TagAutocompleteSelect.vue";

const props = defineProps<{
  publicKey: IPublicKey;
  hasAuthorization?: boolean;
}>();

const emit = defineEmits(["update"]);
const showDialog = ref(false);
const publicKeysStore = usePublicKeysStore();
const snackbar = useSnackbar();

const selectedUsernameOption = ref(FormUsernameOptions.All);
const selectedFilterOption = ref(FormFilterOptions.All);

const filterSelectOptions = [
  { value: FormFilterOptions.All, title: "Allow the key to connect to all available devices" },
  { value: FormFilterOptions.Hostname, title: "Restrict access using a regexp for hostname" },
  { value: FormFilterOptions.Tags, title: "Restrict access by tags" },
];

const usernameSelectOptions = [
  { value: FormUsernameOptions.All, title: "Allow any user" },
  { value: FormUsernameOptions.Username, title: "Restrict access using a regexp for username" },
];

const selectedTags = ref<string[]>([]);
const tagSelectorErrorMessage = ref("");

const {
  value: name,
  errorMessage: nameError,
} = useField<string>("name", yup.string().required());

const {
  value: username,
  errorMessage: usernameError,
} = useField<string>("username", yup.string().required());

const {
  value: hostname,
  errorMessage: hostnameError,
} = useField<string>("hostname", yup.string().required());

const publicKeyData = ref("");

const confirmDisabled = computed(() => {
  if (!name.value || !publicKeyData.value) return true;
  if (selectedFilterOption.value === FormFilterOptions.Tags && selectedTags.value.length === 0) return true;

  return Boolean(
    (!name.value || nameError.value)
    || (selectedUsernameOption.value === FormUsernameOptions.Username && (!username.value || usernameError.value))
    || (selectedFilterOption.value === FormFilterOptions.Hostname && (!hostname.value || hostnameError.value))
    || (selectedFilterOption.value === FormFilterOptions.Tags && !!tagSelectorErrorMessage.value),
  );
});

const populateFilterFields = () => {
  if (!showDialog.value) return;

  const currentUsername = props.publicKey.username;
  selectedUsernameOption.value = currentUsername === ".*" ? FormUsernameOptions.All : FormUsernameOptions.Username;
  username.value = currentUsername;

  const { filter } = props.publicKey;
  if ("tags" in filter && filter.tags.length) {
    selectedTags.value = filter.tags.map((tag) => tag.name);
    selectedFilterOption.value = FormFilterOptions.Tags;
  } else if ("hostname" in filter) {
    const currentHostname = filter.hostname;
    if (currentHostname && currentHostname !== ".*") {
      selectedFilterOption.value = FormFilterOptions.Hostname;
      hostname.value = currentHostname;
    } else if (currentHostname === ".*") {
      selectedFilterOption.value = FormFilterOptions.All;
    }
  }
};

const open = () => {
  showDialog.value = true;
  name.value = props.publicKey.name;
  publicKeyData.value = Buffer.from(props.publicKey.data, "base64").toString("utf-8");
  populateFilterFields();
};

const resetFields = () => {
  hostname.value = "";
  username.value = "";
  selectedTags.value = [];
  tagSelectorErrorMessage.value = "";
};

const close = () => {
  resetFields();
  showDialog.value = false;
};

const constructPublicKey = () => {
  const filterMap = {
    [FormFilterOptions.Hostname]: { hostname: hostname.value?.trim() },
    [FormFilterOptions.Tags]: { tags: selectedTags.value },
    [FormFilterOptions.All]: { hostname: ".*" },
  };

  return {
    ...props.publicKey,
    name: name.value,
    data: props.publicKey.data,
    username: selectedUsernameOption.value === FormUsernameOptions.All ? ".*" : username.value,
    filter: filterMap[selectedFilterOption.value],
  };
};

const edit = async () => {
  if (confirmDisabled.value) return;

  try {
    await publicKeysStore.updatePublicKey(constructPublicKey() as IPublicKey);
    snackbar.showSuccess("Public key updated successfully.");
    emit("update");
    close();
  } catch (error: unknown) {
    snackbar.showError("Failed to update public key.");
    handleError(error);
  }
};

defineExpose({ nameError, usernameError, hostnameError, selectedTags, tagSelectorErrorMessage });
</script>
