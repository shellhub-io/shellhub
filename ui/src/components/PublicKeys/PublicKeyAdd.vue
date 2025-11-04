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
            @click="showDialog = true"
            color="primary"
            tabindex="0"
            variant="elevated"
            aria-label="Add Public Key"
            :disabled="!canCreatePublicKey"
            @keypress.enter="showDialog = true"
            :size="size"
            data-test="public-key-add-btn"
          >
            Add Public Key
          </v-btn>
        </div>
      </template>
      <span>You don't have this kind of authorization.</span>
    </v-tooltip>

    <FormDialog
      v-model="showDialog"
      @close="close"
      @cancel="close"
      @confirm="create"
      title="New Public Key"
      icon="mdi-key-outline"
      confirm-text="Save"
      cancel-text="Cancel"
      :confirm-disabled
      confirm-data-test="pk-add-save-btn"
      cancel-data-test="pk-add-cancel-btn"
      data-test="public-key-add-dialog"
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
            v-model="choiceUsername"
            label="Device username access restriction"
            :items="usernameList"
            item-title="filterText"
            item-value="filterName"
            data-test="username-restriction-field"
          />
        </v-row>

        <v-row class="mt-2 px-3">
          <v-text-field
            v-if="choiceUsername === 'username'"
            v-model="username"
            label="Rule username"
            :error-messages="usernameError"
            data-test="rule-field"
          />
        </v-row>

        <v-row class="mt-4 px-3">
          <v-select
            v-model="choiceFilter"
            label="Device access restriction"
            :items="filterList"
            item-title="filterText"
            item-value="filterName"
            data-test="filter-restriction-field"
          />
        </v-row>

        <v-row class="mt-1 px-3">
          <v-autocomplete
            v-if="choiceFilter === 'tags'"
            v-model="tagChoices"
            v-model:menu="acMenuOpen"
            :menu-props="{ contentClass: menuContentClass, maxHeight: 320 }"
            :items="tags"
            variant="outlined"
            item-title="name"
            item-value="name"
            attach
            chips
            label="Tags"
            density="comfortable"
            :rules="[validateLength]"
            :error-messages="errMsg"
            multiple
            data-test="tags-selector"
            @update:search="onSearch"
          >
            <template #append-item>
              <div ref="sentinel" data-test="tags-sentinel" style="height: 1px;" />
            </template>
          </v-autocomplete>

          <v-text-field
            v-if="choiceFilter === 'hostname'"
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
          :validator="(t) => isKeyValid('public', t)"
          invalid-message="This is not a valid public key."
          @file-name="suggestNameFromFile"
        />

      </div>
    </FormDialog>
  </div>
</template>

<script setup lang="ts">
import { useField } from "vee-validate";
import { computed, nextTick, ref, watch, onMounted, onUnmounted } from "vue";
import * as yup from "yup";
import axios, { AxiosError } from "axios";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import hasPermission from "@/utils/permission";
import { isKeyValid } from "@/utils/sshKeys";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import usePublicKeysStore from "@/store/modules/public_keys";
import { IPublicKeyCreate } from "@/interfaces/IPublicKey";
import useTagsStore from "@/store/modules/tags";
import FileTextComponent from "@/components/Fields/FileTextComponent.vue";

const { size } = defineProps<{ size?: string }>();

type LocalTag = { name: string };
type NameFilterName = "all" | "username";
type DeviceFilterName = "all" | "hostname" | "tags";
interface SelectOption<TName extends string> {
  filterName: TName;
  filterText: string;
}

const emit = defineEmits(["update"]);
const publicKeysStore = usePublicKeysStore();
const tagsStore = useTagsStore();
const showDialog = ref(false);
const snackbar = useSnackbar();

const validateLength = ref(true);
const choiceUsername = ref<NameFilterName>("all");
const choiceFilter = ref<DeviceFilterName>("all");

const filterList = ref<SelectOption<DeviceFilterName>[]>([
  { filterName: "all", filterText: "Allow the key to connect to all available devices" },
  { filterName: "hostname", filterText: "Restrict access using a regexp for hostname" },
  { filterName: "tags", filterText: "Restrict access by tags" },
]);

const usernameList = ref<SelectOption<NameFilterName>[]>([
  { filterName: "all", filterText: "Allow any user" },
  { filterName: "username", filterText: "Restrict access using a regexp for username" },
]);

const tagChoices = ref<string[]>([]);
const errMsg = ref("");
const keyLocal = ref<Record<string, unknown>>({});

const acMenuOpen = ref(false);
const menuContentClass = computed(() => "pk-tags-ac-content");

const fetchedTags = ref<LocalTag[]>([]);
const tags = computed(() => fetchedTags.value);

const sentinel = ref<HTMLElement | null>(null);
const pastedFile = ref<File | null>(null);
let observer: IntersectionObserver | null = null;

const {
  value: name,
  errorMessage: nameError,
  resetField: resetName,
} = useField<string | undefined>("name", yup.string().required(), { initialValue: "" });

const {
  value: username,
  errorMessage: usernameError,
  setErrors: setUsernameError,
  resetField: resetUsername,
} = useField<string | undefined>("username", yup.string().required(), { initialValue: "" });

const {
  value: hostname,
  errorMessage: hostnameError,
  setErrors: setHostnameError,
  resetField: resetHostname,
} = useField<string | undefined>("hostname", yup.string().required(), { initialValue: "" });

const publicKeyData = ref("");
const publicKeyDataError = ref("");

const inputMode = ref<"file" | "text">("file");

const suggestNameFromFile = (filename: string) => {
  if (name.value) return;
  const base = filename.replace(/\.[^.]+$/, "");
  name.value = base || "Imported Public Key";
};

watch([tagChoices, choiceFilter], ([list, currentFilter]) => {
  if (currentFilter !== "tags") {
    validateLength.value = true;
    errMsg.value = "";
    return;
  }
  if (list.length > 3) {
    validateLength.value = false;
    nextTick(() => tagChoices.value.pop());
    errMsg.value = "The maximum capacity has reached";
  } else if (list.length === 0) {
    validateLength.value = false;
    errMsg.value = "You must choose at least one tag";
  } else if (list.length <= 2) {
    validateLength.value = true;
    errMsg.value = "";
  }
});

watch(publicKeyData, () => {
  if (!showDialog.value) return;
  if (!publicKeyData.value) {
    publicKeyDataError.value = "Field is required";
    return;
  }
  if (!isKeyValid("public", publicKeyData.value)) publicKeyDataError.value = "This is not valid key";
  else publicKeyDataError.value = "";
});

const chooseUsername = () => {
  switch (choiceUsername.value) {
    case "all": keyLocal.value = { ...keyLocal.value, username: ".*" }; break;
    case "username": keyLocal.value = { ...keyLocal.value, username: username.value }; break;
    default: break;
  }
};

const chooseFilter = () => {
  switch (choiceFilter.value) {
    case "all":
      keyLocal.value = { ...keyLocal.value, filter: { hostname: ".*" } };
      break;
    case "hostname":
      keyLocal.value = { ...keyLocal.value, filter: { hostname: hostname.value } };
      break;
    case "tags":
      keyLocal.value = { ...keyLocal.value, filter: { tags: tagChoices.value } };
      break;
    default: break;
  }
};

const resetFields = () => {
  resetName();
  resetUsername();
  resetHostname();
  publicKeyData.value = "";
  publicKeyDataError.value = "";
  pastedFile.value = null;
};

const setLocalVariable = () => {
  keyLocal.value = {};
  hostname.value = "";
  tagChoices.value = [];
  choiceFilter.value = "all";
  choiceUsername.value = "all";
  inputMode.value = "file";
  resetFields();
};

watch(showDialog, (open) => {
  if (open) {
    inputMode.value = "file";
  } else {
    setLocalVariable();
  }
});

const close = () => { showDialog.value = false; setLocalVariable(); };

const update = () => { emit("update"); close(); };

const hasErrors = () => {
  if (choiceUsername.value === "username" && !username.value) { setUsernameError("This Field is required!"); return true; }
  if (choiceFilter.value === "hostname" && !hostname.value) { setHostnameError("This Field is required!"); return true; }
  if (choiceFilter.value === "tags" && tagChoices.value.length === 0) return true;
  return false;
};

const create = async () => {
  if (hasErrors()) return;

  try {
    chooseFilter();
    chooseUsername();
    const keySend = {
      ...keyLocal.value,
      data: Buffer.from(publicKeyData.value as string, "utf-8").toString("base64"),
      name: name.value,
    };
    await publicKeysStore.createPublicKey(keySend as IPublicKeyCreate);
    snackbar.showSuccess("Public key created successfully.");
    update();
    resetFields();
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response?.status === 409) {
        publicKeyDataError.value = "Public Key data already exists";
        return;
      }
      snackbar.showError("Failed to create the public key.");
      handleError(error);
    } else {
      snackbar.showError("Failed to create the public key.");
      handleError(error);
    }
  }
};

const page = ref(1);
const perPage = ref(10);
const filter = ref("");
const isLoading = ref(false);

const hasMore = computed(() => tagsStore.numberTags > fetchedTags.value.length);
const canCreatePublicKey = hasPermission("publicKey:create");

const encodeFilter = (search: string) => {
  if (!search) return "";
  const filterToEncodeBase64 = [
    { type: "property", params: { name: "name", operator: "contains", value: search } },
  ];
  return btoa(JSON.stringify(filterToEncodeBase64));
};

const normalizeStoreItems = (arr: unknown[]): LocalTag[] => (arr ?? [])
  .map((tag: unknown) => {
    const name = typeof tag === "string" ? tag : (tag as LocalTag | undefined)?.name;
    return name ? ({ name } as LocalTag) : null;
  })
  .filter((tag: LocalTag | null): tag is LocalTag => !!tag);

const resetPagination = () => {
  page.value = 1;
  perPage.value = 10;
  fetchedTags.value = [];
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

    fetchedTags.value = normalizeStoreItems(tagsStore.list as unknown[]);
  } catch (error) {
    snackbar.showError("Failed to load tags.");
    handleError(error);
  } finally {
    isLoading.value = false;
  }
};

const onSearch = async (search: string) => {
  filter.value = search || "";
  resetPagination();
  await loadTags();
};

watch(choiceFilter, async (val) => {
  if (val === "tags") {
    resetPagination();
    await loadTags();
  }
});

const bumpPerPageAndLoad = async () => {
  if (!hasMore.value || isLoading.value) return;
  perPage.value += 10;
  await loadTags();
};

const getMenuRootEl = (): HTMLElement | null => document.querySelector(`.${menuContentClass.value}`) as HTMLElement | null;

const cleanupObserver = () => {
  if (observer) {
    observer.disconnect();
    observer = null;
  }
};

const setupObserver = () => {
  cleanupObserver();
  const root = getMenuRootEl();
  if (!root || !sentinel.value) return;

  observer = new IntersectionObserver(
    (entries) => {
      const entry = entries[0];
      if (entry?.isIntersecting) bumpPerPageAndLoad();
    },
    { root, threshold: 1.0 },
  );

  observer.observe(sentinel.value);
};

watch(acMenuOpen, async (open) => {
  if (open && choiceFilter.value === "tags") {
    await nextTick();
    setupObserver();
  } else {
    cleanupObserver();
  }
});

onMounted(async () => {
  resetPagination();
  await loadTags();
});

onUnmounted(cleanupObserver);

const confirmDisabled = computed(() => {
  if (!name.value || !publicKeyData.value) return true;

  if (choiceUsername.value === "username" && !username.value) return true;
  if (choiceFilter.value === "hostname" && !hostname.value) return true;
  if (choiceFilter.value === "tags" && tagChoices.value.length === 0) return true;

  const tagRuleBlocking = choiceFilter.value === "tags" && !validateLength.value;

  return Boolean(
    nameError.value
    || publicKeyDataError.value
    || (choiceUsername.value === "username" && usernameError.value)
    || (choiceFilter.value === "hostname" && hostnameError.value)
    || tagRuleBlocking,
  );
});

defineExpose({ publicKeyDataError, nameError, usernameError, hostnameError, errMsg });
</script>
