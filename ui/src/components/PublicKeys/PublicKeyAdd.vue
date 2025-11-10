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
            @update:model-value="handleFilterChange"
          />
        </v-row>

        <v-row class="mt-1 px-3">
          <v-autocomplete
            v-if="selectedFilterOption === FormFilterOptions.Tags"
            v-model="selectedTags"
            v-model:menu="isAutocompleteMenuOpen"
            :menu-props="{ contentClass: menuContentClass, maxHeight: 320 }"
            :items="tags"
            variant="outlined"
            item-title="name"
            item-value="name"
            attach
            chips
            class="mb-4"
            hide-details="auto"
            label="Tags"
            density="comfortable"
            :error-messages="tagSelectorErrorMessage"
            multiple
            data-test="tags-selector"
            @update:model-value="validateSelectedTags"
            @update:search="handleSearch"
          >
            <template #append-item>
              <div
                ref="sentinel"
                data-test="tags-sentinel"
                style="height: 1px"
              />
            </template>
          </v-autocomplete>

          <v-text-field
            v-if="selectedFilterOption === FormFilterOptions.Hostname"
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
import { computed, ref, onMounted } from "vue";
import { useIntersectionObserver } from "@vueuse/core";
import * as yup from "yup";
import axios from "axios";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import hasPermission from "@/utils/permission";
import { isKeyValid } from "@/utils/sshKeys";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import usePublicKeysStore from "@/store/modules/public_keys";
import { IPublicKeyCreate } from "@/interfaces/IPublicKey";
import useTagsStore from "@/store/modules/tags";
import FileTextComponent from "@/components/Fields/FileTextComponent.vue";
import { FormFilterOptions, FormUsernameOptions } from "@/interfaces/IFilter";
import { ITag } from "@/interfaces/ITags";

const { size } = defineProps<{ size?: string }>();

const emit = defineEmits(["update"]);
const publicKeysStore = usePublicKeysStore();
const tagsStore = useTagsStore();
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

const isAutocompleteMenuOpen = ref(false);
const menuContentClass = "pk-tags-ac-content";

const tags = ref<ITag[]>([]);
const hasMoreTagsToLoad = computed(() => tagsStore.numberTags > tags.value.length);
const selectedTags = ref<string[]>([]);
const tagSelectorErrorMessage = ref("");

const sentinel = ref<HTMLElement | null>(null);
const pastedFile = ref<File | null>(null);

const page = ref(1);
const perPage = ref(10);
const filter = ref("");
const isLoading = ref(false);

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

const validateSelectedTags = () => {
  const list = selectedTags.value;
  if (selectedFilterOption.value !== FormFilterOptions.Tags) {
    tagSelectorErrorMessage.value = "";
    return;
  }
  if (list.length > 3) tagSelectorErrorMessage.value = "You can select up to three tags only";
  else if (list.length === 0) tagSelectorErrorMessage.value = "You must choose at least one tag";
  else if (list.length <= 3) tagSelectorErrorMessage.value = "";
};

const handleFilterChange = async () => {
  if (selectedFilterOption.value === FormFilterOptions.Tags) await loadTags();
  else tagSelectorErrorMessage.value = "";
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

const encodeFilter = (filterQuery: string) => {
  if (!filterQuery) return "";
  const filterToEncodeBase64 = [{ type: "property", params: { name: "name", operator: "contains", value: filterQuery } }];
  return btoa(JSON.stringify(filterToEncodeBase64));
};

const resetPagination = () => {
  page.value = 1;
  perPage.value = 10;
};

const loadTags = async () => {
  if (isLoading.value) return;
  isLoading.value = true;

  try {
    await tagsStore.autocomplete({
      tenant: localStorage.getItem("tenant") || "",
      filter: encodeFilter(filter.value),
      page: page.value,
      perPage: perPage.value,
    });
    tags.value = tagsStore.list;
  } catch (error) {
    snackbar.showError("Failed to load tags.");
    handleError(error);
  } finally {
    isLoading.value = false;
  }
};

const handleSearch = async (filterQuery = "") => {
  filter.value = filterQuery;
  resetPagination();
  await loadTags();
};

const bumpPerPageAndLoad = async () => {
  if (!hasMoreTagsToLoad.value || isLoading.value) return;
  perPage.value += 10;
  await loadTags();
};

const getMenuRootEl = (): HTMLElement | null => document.querySelector(`.${menuContentClass}`);

useIntersectionObserver(
  sentinel,
  ([{ isIntersecting }]) => { if (isIntersecting) void bumpPerPageAndLoad(); },
  { root: getMenuRootEl, threshold: 1.0 },
);

onMounted(async () => {
  resetPagination();
  await loadTags();
});

defineExpose({ nameError, usernameError, hostnameError });
</script>
