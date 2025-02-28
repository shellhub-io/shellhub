<template>
  <div>
    <v-tooltip v-bind="$attrs" class="text-center" location="bottom" :disabled="canCreatePublicKey">
      <template v-slot:activator="{ props }">
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
      <span> You don't have this kind of authorization. </span>
    </v-tooltip>

    <BaseDialog v-model="showDialog" @close="close" transition="dialog-bottom-transition">
      <v-card class="bg-v-theme-surface">
        <v-card-title class="text-h5 pa-3 bg-primary" data-test="pk-add-title">
          New Public Key
        </v-card-title>
        <form @submit.prevent="create" class="mt-3">
          <v-card-text>
            <v-text-field
              v-model="name"
              :error-messages="nameError"
              label="Name"
              placeholder="Name used to identify the public key"
              data-test="name-field"
            />

            <v-row class="mt-1 px-3">
              <v-select
                v-model="choiceUsername"
                label="Device username access restriction"
                :items="usernameList"
                item-title="filterText"
                item-value="filterName"
                data-test="username-restriction-field"
              />
            </v-row>

            <v-text-field
              v-if="choiceUsername === 'username'"
              v-model="username"
              label="Rule username"
              :error-messages="usernameError"
              data-test="rule-field"
            />

            <v-row class="mt-1 px-3">
              <v-select
                v-model="choiceFilter"
                label="Device access restriction"
                :items="filterList"
                item-title="filterText"
                item-value="filterName"
                data-test="filter-restriction-field"
              />
            </v-row>

            <v-row class="px-3">
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
                :rules="[validateLength]"
                :error-messages="errMsg"
                multiple
                data-test="tags-selector"
                @update:search="onSearch"
              >=
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

            <v-textarea
              v-model="publicKeyData"
              class="mt-5"
              label="Public key data"
              :error-messages="publicKeyDataError"
              required
              messages="Supports RSA, DSA, ECDSA (NIST P-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats."
              data-test="data-field"
              rows="2"
            />
          </v-card-text>

          <v-card-actions>
            <v-spacer />
            <v-btn @click="close" data-test="pk-add-cancel-btn">Cancel</v-btn>
            <v-btn color="primary" type="submit" data-test="pk-add-save-btn">Save</v-btn>
          </v-card-actions>
        </form>
      </v-card>
    </BaseDialog>
  </div>
</template>

<script setup lang="ts">
import { useField } from "vee-validate";
import { computed, nextTick, ref, watch, onMounted, onUnmounted } from "vue";
import * as yup from "yup";
import axios, { AxiosError } from "axios";
import hasPermission from "@/utils/permission";
import { isKeyValid } from "@/utils/sshKeys";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
import usePublicKeysStore from "@/store/modules/public_keys";
import { IPublicKeyCreate } from "@/interfaces/IPublicKey";
import useTagsStore from "@/store/modules/tags";

const { size } = defineProps<{ size?: string }>();

const emit = defineEmits(["update"]);
const publicKeysStore = usePublicKeysStore();
const tagsStore = useTagsStore();
const showDialog = ref(false);
const snackbar = useSnackbar();

const validateLength = ref(true);
const choiceFilter = ref<"all" | "hostname" | "tags">("all");
const choiceUsername = ref<"all" | "username">("all");
const tagChoices = ref<string[]>([]);
const errMsg = ref("");
const keyLocal = ref({});

const usernameList = ref([
  { filterName: "all", filterText: "Allow any user" },
  { filterName: "username", filterText: "Restrict access using a regexp for username" },
]);

const filterList = ref([
  { filterName: "all", filterText: "Allow the key to connect to all available devices" },
  { filterName: "hostname", filterText: "Restrict access using a regexp for hostname" },
  { filterName: "tags", filterText: "Restrict access by tags" },
]);

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

const {
  value: publicKeyData,
  errorMessage: publicKeyDataError,
  setErrors: setPublicKeyDataError,
  resetField: resetPublicKeyData,
} = useField<string>("publicKeyData", yup.string().required(), { initialValue: "" });

type LocalTag = { name: string };

const acMenuOpen = ref(false);
const menuContentClass = computed(() => "pk-tags-ac-content");

const fetchedTags = ref<LocalTag[]>([]);
const tags = computed(() => fetchedTags.value);

const sentinel = ref<HTMLElement | null>(null);
let observer: IntersectionObserver | null = null;

const page = ref(1);
const perPage = ref(10);
const filter = ref("");
const isLoading = ref(false);

const hasMore = computed(() => tagsStore.numberTags > fetchedTags.value.length);

const canCreatePublicKey = hasPermission("publicKey:create");

watch(tagChoices, (list) => {
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

watch(publicKeyData, async () => {
  if (publicKeyData.value !== "") setPublicKeyDataError("Field is required");
  if (isKeyValid("public", publicKeyData.value)) setPublicKeyDataError("This is not valid key");
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

const setLocalVariable = () => {
  keyLocal.value = {};
  hostname.value = "";
  tagChoices.value = [];
  choiceFilter.value = "all";
  choiceUsername.value = "all";
};

watch(showDialog, (value) => { if (!value) setLocalVariable(); });

const close = () => { showDialog.value = false; setLocalVariable(); };
const update = () => { emit("update"); close(); };

const hasErrors = () => {
  if (choiceUsername.value === "username" && !username.value) { setUsernameError("This Field is required!"); return true; }
  if (choiceFilter.value === "hostname" && !hostname.value) { setHostnameError("This Field is required!"); return true; }
  if (choiceFilter.value === "tags" && tagChoices.value.length === 0) return true;
  return false;
};

const resetFields = () => { resetName(); resetUsername(); resetHostname(); resetPublicKeyData(); };

const create = async () => {
  if (hasErrors()) return;

  try {
    chooseFilter();
    chooseUsername();
    const keySend = {
      ...keyLocal.value,
      data: btoa(publicKeyData.value),
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
        setPublicKeyDataError("Public Key data already exists");
      }
    } else {
      snackbar.showError("Failed to create the public key.");
      handleError(error);
    }
  }
};

const encodeFilter = (search: string) => {
  if (!search) return "";
  const filterToEncodeBase64 = [
    { type: "property", params: { name: "name", operator: "contains", value: search } },
  ];
  return btoa(JSON.stringify(filterToEncodeBase64));
};

const normalizeStoreItems = (arr): LocalTag[] => (arr ?? [])
  .map((tag) => {
    const name = typeof tag === "string" ? tag : tag?.name;
    return name ? ({ name } as LocalTag) : null;
  })
  .filter((t: LocalTag | null): t is LocalTag => !!t);

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

    fetchedTags.value = normalizeStoreItems(tagsStore.list);
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

defineExpose({ publicKeyDataError, nameError });
</script>
