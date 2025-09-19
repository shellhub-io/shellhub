<template>
  <div>
    <v-list-item
      @click="open()"
      v-bind="$attrs"
      :disabled="!hasAuthorization"
      data-test="public-key-edit-title-btn"
    >
      <div class="d-flex align-center">
        <div data-test="public-key-edit-icon" class="mr-2">
          <v-icon> mdi-pencil </v-icon>
        </div>

        <v-list-item-title>
          Edit
        </v-list-item-title>
      </div>
    </v-list-item>

    <BaseDialog v-model="showDialog" @close="close" transition="dialog-bottom-transition">
      <v-card class="bg-v-theme-surface">
        <v-card-title class="text-h5 pa-3 bg-primary" data-test="public-key-edit-title">
          Edit Public Key
        </v-card-title>
        <form @submit.prevent="edit" class="mt-3">
          <v-card-text>
            <v-text-field
              v-model="name"
              label="Key name"
              placeholder="Name used to identify the public key"
              :error-messages="nameError"
              required
              variant="underlined"
              data-test="name-field"
            />

            <v-row class="mt-1 px-3">
              <v-select
                v-model="choiceUsername"
                label="Device username access restriction"
                :items="usernameList"
                variant="underlined"
                item-title="filterText"
                item-value="filterName"
                data-test="username-restriction-field"
              />
            </v-row>

            <v-text-field
              v-if="choiceUsername === 'username'"
              v-model="username"
              label="Rule username"
              variant="underlined"
              :error-messages="usernameError"
              data-test="rule-field"
            />

            <v-row class="mt-1 px-3">
              <v-select
                v-model="choiceFilter"
                label="Device access restriction"
                :items="filterList"
                variant="underlined"
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
            <v-btn
              @click="close"
              data-test="pk-edit-cancel-btn"
            >
              Cancel
            </v-btn>
            <v-btn
              color="primary"
              type="submit"
              data-test="pk-edit-save-btn"
            >
              Save
            </v-btn>
          </v-card-actions>
        </form>
      </v-card>
    </BaseDialog>
  </div>
</template>

<script setup lang="ts">
import { useField } from "vee-validate";
import {
  ref,
  watch,
  onMounted,
  computed,
  nextTick,
  onUpdated,
  onUnmounted,
} from "vue";
import * as yup from "yup";
import { IPublicKey } from "@/interfaces/IPublicKey";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
import { HostnameFilter, TagsFilter } from "@/interfaces/IFilter";
import usePublicKeysStore from "@/store/modules/public_keys";
import useTagsStore from "@/store/modules/tags";
import { ITag } from "@/interfaces/ITags";

const props = defineProps<{
  publicKey: IPublicKey;
  hasAuthorization?: boolean;
}>();

const emit = defineEmits(["update"]);
const showDialog = ref(false);
const publicKeysStore = usePublicKeysStore();
const tagsStore = useTagsStore();
const snackbar = useSnackbar();
const choiceFilter = ref("hostname");
const validateLength = ref(true);
const errMsg = ref("");
const prop = computed(() => props);
const choiceUsername = ref("username");

const filterList = ref([
  { filterName: "all", filterText: "Allow the key to connect to all available devices" },
  { filterName: "hostname", filterText: "Restrict access using a regexp for hostname" },
  { filterName: "tags", filterText: "Restrict access by tags" },
]);

const usernameList = ref([
  { filterName: "all", filterText: "Allow any user" },
  { filterName: "username", filterText: "Restrict access using a regexp for username" },
]);

const tagChoices = ref<ITag[]>([]);
const keyLocal = ref<Partial<IPublicKey>>({ name: "", username: "", data: "" });

const {
  value: name,
  errorMessage: nameError,
} = useField<string>("name", yup.string().required(), {
  initialValue: prop.value.publicKey.name,
});
watch(name, () => { keyLocal.value.name = name.value; });

const {
  value: username,
  errorMessage: usernameError,
  setErrors: setUsernameError,
} = useField<string>("username", yup.string().required(), {
  initialValue: prop.value.publicKey.username,
});
watch(username, () => { keyLocal.value.username = username.value; });

const {
  value: hostname,
  errorMessage: hostnameError,
  setErrors: setHostnameError,
} = useField<string>("hostname", yup.string().required(), {
  initialValue: (prop.value.publicKey.filter as HostnameFilter)?.hostname || "",
});

const {
  value: publicKeyData,
  errorMessage: publicKeyDataError,
} = useField<string>("publicKeyData", yup.string().required(), {
  initialValue: prop.value.publicKey.data,
});

const hasAuthorization = computed(() => props.hasAuthorization ?? true);

const hasTags = computed(() => {
  const { publicKey } = props;
  if (!publicKey) return false;
  return Reflect.ownKeys(publicKey.filter)[0] === "tags";
});

type LocalTag = { name: string };

const acMenuOpen = ref(false);
const menuContentClass = computed(() => `pk-edit-tags-ac-${(props.publicKey?.name || "key").replace(/\W/g, "-")}`);

const fetchedTags = ref<LocalTag[]>([]);
const tags = computed(() => fetchedTags.value);

const sentinel = ref<HTMLElement | null>(null);
let observer: IntersectionObserver | null = null;

const page = ref(1);
const perPage = ref(10);
const filter = ref("");
const isLoading = ref(false);

const hasMore = computed(() => tagsStore.numberTags > fetchedTags.value.length);

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

watch(choiceFilter, async (val) => {
  if (val === "tags") {
    resetPagination();
    await loadTags();
  }
});

watch(tagChoices, (list) => {
  if (list.length > 3) {
    validateLength.value = false;
    nextTick(() => tagChoices.value.pop());
    errMsg.value = "The maximum capacity has reached";
  } else if (list.length <= 2) {
    validateLength.value = true;
    errMsg.value = "";
  }
});

const handleUpdate = () => {
  if (showDialog.value) {
    if (hasTags.value) {
      const { tags } = props.publicKey.filter as TagsFilter;
      tagChoices.value = tags;
      choiceFilter.value = "tags";
    } else {
      const { hostname: hostnameLocal } = props.publicKey.filter as HostnameFilter;
      if (!!hostnameLocal && hostnameLocal !== ".*") {
        choiceFilter.value = "hostname";
        hostname.value = hostnameLocal;
      } else if (!!hostnameLocal && hostnameLocal === ".*") {
        choiceFilter.value = "all";
      }
    }

    const { username: usernameLocal } = props.publicKey;
    choiceUsername.value = usernameLocal === ".*" ? "all" : "username";
    username.value = usernameLocal;
  }
};

const chooseFilter = () => {
  switch (choiceFilter.value) {
    case "all": {
      keyLocal.value = { ...keyLocal.value, filter: { hostname: ".*" } };
      break;
    }
    case "hostname": {
      keyLocal.value = { ...keyLocal.value, filter: { hostname: hostname.value } };
      break;
    }
    case "tags": {
      keyLocal.value = { ...keyLocal.value, filter: { tags: tagChoices.value } };
      break;
    }
    default: {
      break;
    }
  }
};

const chooseUsername = () => {
  switch (choiceUsername.value) {
    case "all": {
      keyLocal.value = { ...keyLocal.value, username: ".*" };
      break;
    }
    case "username": {
      keyLocal.value = { ...keyLocal.value, username: username.value };
      break;
    }
    default: {
      break;
    }
  }
};

const setLocalVariable = () => {
  keyLocal.value = { ...props.publicKey };
  keyLocal.value.data = atob(props.publicKey.data);
};

const hasError = () => {
  if (choiceUsername.value === "username" && username.value === "") {
    setUsernameError("This Field is required !");
    return true;
  }
  if (choiceFilter.value === "hostname" && hostname.value === "") {
    setHostnameError("This Field is required !");
    return true;
  }
  if (choiceFilter.value === "tags" && tagChoices.value.length === 0) {
    return true;
  }
  return false;
};

const open = () => {
  showDialog.value = true;
  name.value = props.publicKey.name;
  publicKeyData.value = props.publicKey.data;
};

onMounted(() => {
  setLocalVariable();
  resetPagination();
  loadTags();
});

onUpdated(() => {
  handleUpdate();
  setLocalVariable();
  keyLocal.value.data = publicKeyData.value;
});

onUnmounted(() => {
  cleanupObserver();
});

const resetPublicKey = () => {
  hostname.value = "";
  username.value = "";
  tagChoices.value = [];
};

const close = () => {
  resetPublicKey();
  setLocalVariable();
  showDialog.value = false;
  cleanupObserver();
};

const update = () => {
  emit("update");
  close();
};

const edit = async () => {
  if (!hasError()) {
    chooseFilter();
    chooseUsername();
    const keySend = { ...keyLocal.value, data: btoa(keyLocal.value.data as string) };

    try {
      await publicKeysStore.updatePublicKey(keySend as IPublicKey);
      snackbar.showSuccess("Public key updated successfully.");
      update();
    } catch (error: unknown) {
      snackbar.showError("Failed to update public key.");
      handleError(error);
    }
  }
};

defineExpose({ nameError, usernameError, hostnameError });
</script>
