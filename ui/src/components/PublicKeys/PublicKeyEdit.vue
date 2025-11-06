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
          <v-icon> mdi-pencil </v-icon>
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
            item-title="name"
            item-value="name"
            attach
            chips
            label="Tags"
            :rules="[validateLength]"
            :error-messages="errMsg"
            :messages="noTagsSelected ? 'No tags selected' : ''"
            placeholder="Select up to 3 tags"
            variant="outlined"
            density="comfortable"
            multiple
            data-test="tags-selector"
            @update:search="onSearch"
          >
            <template #append-item>
              <div
                ref="sentinel"
                data-test="tags-sentinel"
                style="height: 1px;"
              />
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
import { ref, watch, onMounted, computed, nextTick, onUnmounted } from "vue";
import * as yup from "yup";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import { IPublicKey } from "@/interfaces/IPublicKey";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import { HostnameFilter, TagsFilter } from "@/interfaces/IFilter";
import usePublicKeysStore from "@/store/modules/public_keys";
import useTagsStore from "@/store/modules/tags";
import FileTextComponent from "@/components/Fields/FileTextComponent.vue";
import { ITag } from "@/interfaces/ITags";

type TagName = { name: string };
type TagsFilterNames = { tags: string[] };
type LocalFilter = HostnameFilter | TagsFilterNames;
type LocalPublicKey = Omit<IPublicKey, "filter"> & { filter: LocalFilter };

const props = defineProps<{
  publicKey: IPublicKey;
  hasAuthorization?: boolean;
}>();

const emit = defineEmits(["update"]);
const showDialog = ref(false);
const publicKeysStore = usePublicKeysStore();
const tagsStore = useTagsStore();
const snackbar = useSnackbar();

const choiceFilter = ref<"all" | "hostname" | "tags">("hostname");
const choiceUsername = ref<"all" | "username">("username");

const validateLength = ref(true);
const errMsg = ref("");

const filterList = ref([
  { filterName: "all", filterText: "Allow the key to connect to all available devices" },
  { filterName: "hostname", filterText: "Restrict access using a regexp for hostname" },
  { filterName: "tags", filterText: "Restrict access by tags" },
]);

const usernameList = ref([
  { filterName: "all", filterText: "Allow any user" },
  { filterName: "username", filterText: "Restrict access using a regexp for username" },
]);

const tagChoices = ref<string[]>([]);
const keyLocal = ref<Partial<LocalPublicKey>>({ name: "", username: "", data: "" });

const {
  value: name,
  errorMessage: nameError,
} = useField<string>("name", yup.string().required(), {
  initialValue: props.publicKey.name,
});
watch(name, () => { keyLocal.value.name = name.value; });

const {
  value: username,
  errorMessage: usernameError,
} = useField<string>("username", yup.string().required(), {
  initialValue: props.publicKey.username,
});
watch(username, () => { keyLocal.value.username = username.value; });

const {
  value: hostname,
  errorMessage: hostnameError,
} = useField<string>("hostname", yup.string().required(), {
  initialValue: (props.publicKey.filter as HostnameFilter)?.hostname || "",
});

const publicKeyData = ref("");

const hasAuthorization = computed(() => props.hasAuthorization ?? true);

const hasTags = computed(() => {
  const { publicKey } = props;
  if (!publicKey) return false;
  return Reflect.ownKeys(publicKey.filter)[0] === "tags";
});

const acMenuOpen = ref(false);
const menuContentClass = computed(
  () => `pk-edit-tags-ac-${(props.publicKey?.name || "key").replace(/\W/g, "-")}`,
);

const fetchedTags = ref<TagName[]>([]);
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
  return Buffer.from(JSON.stringify(filterToEncodeBase64), "utf-8").toString("base64");
};

const normalizeStoreItems = (arr: ITag[]): TagName[] => (arr ?? [])
  .map((tag) => {
    const name = typeof tag === "string" ? tag : tag?.name;
    return name ? ({ name } as TagName) : null;
  })
  .filter((tag: TagName | null): tag is TagName => !!tag);

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

const getMenuRootEl = (): HTMLElement | null => document.querySelector(`.${menuContentClass.value}`);

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
      if (entry?.isIntersecting) void bumpPerPageAndLoad();
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
  } else {
    acMenuOpen.value = false;
  }
});

watch([tagChoices, choiceFilter], async ([list, filterMode]) => {
  if (filterMode !== "tags") {
    validateLength.value = true;
    errMsg.value = "";
    return;
  }
  if (list.length > 3) {
    validateLength.value = false;
    await nextTick(() => tagChoices.value.pop());
    errMsg.value = "The maximum capacity has reached";
  } else {
    validateLength.value = true;
    errMsg.value = "";
  }
});

const toTagNames = (tagsIn: unknown): string[] => {
  if (!Array.isArray(tagsIn)) return [];
  return (tagsIn as Array<string | { name?: string }>)
    .map((t) => (typeof t === "string" ? t : t?.name))
    .filter((n): n is string => !!n);
};

const noTagsSelected = computed(
  () => choiceFilter.value === "tags" && toTagNames(tagChoices.value).length === 0,
);

const handleUpdate = () => {
  if (!showDialog.value) return;

  if (hasTags.value) {
    const { tags } = props.publicKey.filter as TagsFilter;
    tagChoices.value = toTagNames(tags);
    choiceFilter.value = "tags";
  } else {
    const { hostname: hostnameLocal } = props.publicKey.filter as HostnameFilter;
    if (hostnameLocal && hostnameLocal !== ".*") {
      choiceFilter.value = "hostname";
      hostname.value = hostnameLocal;
    } else if (hostnameLocal === ".*") {
      choiceFilter.value = "all";
    }
  }

  const { username: usernameLocal } = props.publicKey;
  choiceUsername.value = usernameLocal === ".*" ? "all" : "username";
  username.value = usernameLocal;
};

const setLocalVariable = () => {
  keyLocal.value = { ...(props.publicKey as LocalPublicKey) };
};

const open = () => {
  showDialog.value = true;
  name.value = props.publicKey.name;
  publicKeyData.value = Buffer.from(props.publicKey.data, "base64").toString("utf-8");
  handleUpdate();
};

onMounted(async () => {
  setLocalVariable();
  resetPagination();
  await loadTags();
});

onUnmounted(() => {
  cleanupObserver();
});

const resetPublicKey = () => {
  hostname.value = "";
  username.value = "";
  tagChoices.value = [];
  validateLength.value = true;
  errMsg.value = "";
  acMenuOpen.value = false;
  cleanupObserver();
  page.value = 1;
  perPage.value = 10;
  filter.value = "";
  fetchedTags.value = [];
};

const close = () => {
  resetPublicKey();
  setLocalVariable();
  showDialog.value = false;
};

const update = () => {
  emit("update");
  close();
};

const isUsernameMissing = computed(() => choiceUsername.value === "username" && username.value.trim() === "");

const isHostnameMissing = computed(() => choiceFilter.value === "hostname" && hostname.value.trim() === "");

const areTagsMissing = computed(() => choiceFilter.value === "tags" && toTagNames(tagChoices.value).length === 0);

const confirmDisabled = computed(() => {
  if (!name.value) return true;

  const tagRuleBlocking = choiceFilter.value === "tags" && !validateLength.value;

  return Boolean(
    nameError.value
    || isUsernameMissing.value
    || isHostnameMissing.value
    || areTagsMissing.value
    || tagRuleBlocking,
  );
});

const edit = async () => {
  if (confirmDisabled.value) return;

  const usernameToSend = choiceUsername.value === "all" ? ".*" : username.value;

  let filterToSend: LocalFilter;
  if (choiceFilter.value === "all") {
    filterToSend = { hostname: ".*" };
  } else if (choiceFilter.value === "hostname") {
    filterToSend = { hostname: hostname.value };
  } else {
    filterToSend = { tags: toTagNames(tagChoices.value) };
  }

  const keySend = {
    ...(keyLocal.value as LocalPublicKey),
    username: usernameToSend,
    filter: filterToSend,
  };

  try {
    await publicKeysStore.updatePublicKey(keySend as unknown as IPublicKey);
    snackbar.showSuccess("Public key updated successfully.");
    update();
  } catch (error: unknown) {
    snackbar.showError("Failed to update public key.");
    handleError(error);
  }
};

defineExpose({ nameError, usernameError, hostnameError, errMsg });
</script>
