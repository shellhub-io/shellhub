<template>
  <div>
    <v-tooltip
      v-bind="$attrs"
      class="text-center"
      location="bottom"
      :disabled="canCreatePublicKey"
    >
      <template #activator="{ props }">
        <div v-bind="props">
          <v-btn
            color="primary"
            tabindex="0"
            variant="elevated"
            aria-label="Add Public Key"
            :disabled="!canCreatePublicKey"
            :size="size"
            data-test="public-key-add-btn"
            @click="showDialog = true"
            @keypress.enter="showDialog = true"
          >
            Add Public Key
          </v-btn>
        </div>
      </template>
      <span>You don't have this kind of authorization.</span>
    </v-tooltip>

    <FormDialog
      v-model="showDialog"
      title="New Public Key"
      icon="mdi-key-outline"
      confirm-text="Save"
      cancel-text="Cancel"
      :confirm-disabled
      confirm-data-test="pk-add-save-btn"
      cancel-data-test="pk-add-cancel-btn"
      data-test="public-key-add-dialog"
      @close="close"
      @cancel="close"
      @confirm="create"
    >
      <div class="px-6 pt-4">
        <v-row class="mt-1 px-3">
          <v-text-field
            v-model="name"
            :error-messages="nameError"
            label="Name"
            placeholder="Name used to identify the public key"
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
          v-model:error-message="publicKeyDataError"
          class="mt-4 mb-2"
          enable-paste
          :pasted-file
          textarea-label="Public key data"
          description-text="Supports RSA, DSA, ECDSA (NIST P-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats."
          :validator="(key: string) => isKeyValid('public', key)"
          invalid-message="This is not a valid public key."
          @file-name="suggestNameFromFile"
          @update:model-value="handlePublicKeyDataChange"
        />
      </div>
    </FormDialog>
  </div>
</template>

<script setup lang="ts">
import { useField } from "vee-validate";
import { computed, ref } from "vue";
import * as yup from "yup";
import axios from "axios";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import hasPermission from "@/utils/permission";
import { isKeyValid } from "@/utils/sshKeys";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import usePublicKeysStore from "@/store/modules/public_keys";
import { IPublicKeyCreate } from "@/interfaces/IPublicKey";
import FileTextComponent from "@/components/Fields/FileTextComponent.vue";
import { FormFilterOptions, FormUsernameOptions } from "@/interfaces/IFilter";
import TagAutocompleteSelect from "@/components/Tags/TagAutocompleteSelect.vue";

const { size } = defineProps<{ size?: string }>();

const emit = defineEmits(["update"]);
const publicKeysStore = usePublicKeysStore();
const showDialog = ref(false);
const snackbar = useSnackbar();

const selectedUsernameOption = ref(FormUsernameOptions.All);
const selectedFilterOption = ref(FormFilterOptions.All);

const filterSelectOptions = [
  { value: "all", title: "Allow the key to connect to all available devices" },
  { value: "hostname", title: "Restrict access using a regexp for hostname" },
  { value: "tags", title: "Restrict access by tags" },
];

const usernameSelectOptions = [
  { value: "all", title: "Allow any user" },
  { value: "username", title: "Restrict access using a regexp for username" },
];

const selectedTags = ref<string[]>([]);
const tagSelectorErrorMessage = ref("");

const pastedFile = ref<File | null>(null);

const {
  value: name,
  errorMessage: nameError,
  resetField: resetName,
} = useField<string>("name", yup.string().required(), { initialValue: "" });

const {
  value: username,
  errorMessage: usernameError,
  resetField: resetUsername,
} = useField<string>("username", yup.string().required(), { initialValue: "" });

const {
  value: hostname,
  errorMessage: hostnameError,
  resetField: resetHostname,
} = useField<string>("hostname", yup.string().required(), { initialValue: "" });

const publicKeyData = ref("");
const publicKeyDataError = ref("");
const inputMode = ref<"file" | "text">("file");

const confirmDisabled = computed(() => {
  if (!name.value || !publicKeyData.value) return true;
  if (selectedFilterOption.value === FormFilterOptions.Tags && selectedTags.value.length === 0) return true;

  return Boolean(
    (!name.value || nameError.value)
    || (!publicKeyData.value || publicKeyDataError.value)
    || (selectedUsernameOption.value === FormUsernameOptions.Username && (!username.value || usernameError.value))
    || (selectedFilterOption.value === FormFilterOptions.Hostname && (!hostname.value || hostnameError.value))
    || (selectedFilterOption.value === FormFilterOptions.Tags && !!tagSelectorErrorMessage.value),
  );
});

const canCreatePublicKey = hasPermission("publicKey:create");

const suggestNameFromFile = (filename: string) => {
  if (name.value) return;
  const base = filename.replace(/\.[^.]+$/, "");
  name.value = base || "Imported Public Key";
};

const handlePublicKeyDataChange = () => {
  if (!publicKeyData.value) {
    publicKeyDataError.value = "Field is required";
    return;
  }
  if (!isKeyValid("public", publicKeyData.value)) publicKeyDataError.value = "This is not a valid key";
  else publicKeyDataError.value = "";
};

const resetFields = () => {
  resetName();
  resetUsername();
  resetHostname();
  publicKeyData.value = "";
  publicKeyDataError.value = "";
  selectedFilterOption.value = FormFilterOptions.All;
  selectedUsernameOption.value = FormUsernameOptions.All;
  selectedTags.value = [];
  inputMode.value = "file";
  pastedFile.value = null;
};

const close = () => {
  showDialog.value = false;
  resetFields();
};

const constructNewPublicKey = () => {
  const filterMap = {
    [FormFilterOptions.Hostname]: { hostname: hostname.value.trim() },
    [FormFilterOptions.Tags]: { tags: selectedTags.value },
    [FormFilterOptions.All]: { hostname: ".*" },
  };

  return {
    data: Buffer.from(publicKeyData.value, "utf-8").toString("base64"),
    name: name.value,
    username: selectedUsernameOption.value === FormUsernameOptions.All ? ".*" : username.value?.trim(),
    filter: filterMap[selectedFilterOption.value],
  };
};

const create = async () => {
  try {
    await publicKeysStore.createPublicKey(constructNewPublicKey() as IPublicKeyCreate);
    snackbar.showSuccess("Public key created successfully.");
    emit("update");
    close();
    resetFields();
  } catch (error: unknown) {
    if (axios.isAxiosError(error) && error.response?.status === 409) {
      publicKeyDataError.value = "Public Key data already exists";
      return;
    }
    snackbar.showError("Failed to create the public key.");
    handleError(error);
  }
};

defineExpose({ nameError, usernameError, hostnameError });
</script>
