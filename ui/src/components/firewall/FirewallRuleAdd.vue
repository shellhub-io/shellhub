<template>
  <div>
    <v-tooltip v-bind="$attrs" class="text-center" location="bottom" :disabled="canCreateFirewallRule">
      <template v-slot:activator="{ props }">
        <div v-bind="props">
          <v-btn
            v-bind="$attrs"
            @click="open"
            color="primary"
            tabindex="0"
            variant="elevated"
            :disabled="!canCreateFirewallRule"
            data-test="firewall-add-rule-btn"
          >
            Add Rule
          </v-btn>
        </div>
      </template>
      <span> You don't have this kind of authorization. </span>
    </v-tooltip>

    <FormDialog
      v-model="showDialog"
      @close="close"
      @cancel="close"
      @confirm="addFirewallRule"
      title="New Firewall Rule"
      icon="mdi-shield-check"
      confirm-text="Add"
      cancel-text="Cancel"
      :confirm-disabled="hasErrors"
      confirm-data-test="firewall-rule-add-btn"
      cancel-data-test="firewall-rule-cancel"
      data-test="firewall-rule-dialog"
    >
      <div class="px-6 pt-4">
        <v-row>
          <v-col>
            <v-select
              v-model="active"
              :items="activeSelectOptions"
              label="Rule status"
              data-test="firewall-rule-status"
            />
          </v-col>

          <v-col>
            <v-text-field
              v-model="priority"
              label="Rule priority"
              :error-messages="priorityError"
              type="number"
              data-test="firewall-rule-priority"
            />
          </v-col>

          <v-col>
            <v-select
              v-model="action"
              :items="actionSelectOptions"
              label="Rule policy"
              data-test="firewall-rule-policy"
            />
          </v-col>
        </v-row>

        <v-row class="mt-1 mb-1 px-3">
          <v-select
            v-model="selectedIPOption"
            @update:model-value="handleSourceIpUpdate"
            label="Source IP access restriction"
            :items="sourceIPSelectOptions"
            data-test="firewall-rule-source-ip-select"
          />
        </v-row>

        <v-text-field
          v-if="selectedIPOption === 'restrict'"
          v-model="sourceIp"
          label="Rule source IP"
          :error-messages="sourceIpError"
          data-test="firewall-rule-source-ip"
        />

        <v-row class="mt-1 mb-1 px-3">
          <v-select
            v-model="selectedUsernameOption"
            @update:model-value="handleUsernameUpdate"
            label="Device username access restriction"
            :items="usernameSelectOptions"
            data-test="username-field"
          />
        </v-row>

        <v-text-field
          v-if="selectedUsernameOption === 'username'"
          v-model="username"
          label="Username access restriction"
          placeholder="Username used during the connection"
          :error-messages="usernameError"
          data-test="firewall-rule-username-restriction"
        />

        <v-row class="mt-2 mb-1 px-3">
          <v-select
            v-model="selectedFilterOption"
            @update:model-value="handleFilterUpdate"
            label="Device access restriction"
            :items="filterSelectOptions"
            data-test="filter-select"
          />
        </v-row>

        <v-text-field
          v-if="selectedFilterOption === FormFilterOptions.Hostname"
          v-model="hostname"
          label="Device hostname access restriction"
          placeholder="Device hostname used during the connection"
          :error-messages="hostnameError"
          data-test="firewall-rule-hostname-restriction"
        />

        <v-row v-else-if="selectedFilterOption === FormFilterOptions.Tags" class="px-3 mt-2">
          <v-autocomplete
            v-model="selectedTags"
            v-model:menu="acMenuOpen"
            :menu-props="{ contentClass: menuContentClass, maxHeight: 320 }"
            :items="tags"
            item-title="name"
            item-value="name"
            attach
            chips
            label="Tags"
            :error-messages="selectedTagsError"
            variant="outlined"
            multiple
            data-test="tags-selector"
            @update:model-value="setSelectedTagsError"
            @update:search="onSearch"
          >
            <template #append-item>
              <div ref="sentinel" data-test="tags-sentinel" style="height: 1px;" />
            </template>
          </v-autocomplete>
        </v-row>
      </div>
    </FormDialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, nextTick, onUnmounted, watch } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import { IFirewallRule } from "@/interfaces/IFirewallRule";
import hasPermission from "@/utils/permission";
import { envVariables } from "@/envVariables";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import { FormFilterOptions } from "@/interfaces/IFilter";
import useFirewallRulesStore from "@/store/modules/firewall_rules";
import useTagsStore from "@/store/modules/tags";
import useUsersStore from "@/store/modules/users";

type LocalTag = { name: string };

const snackbar = useSnackbar();
const firewallRulesStore = useFirewallRulesStore();
const tagsStore = useTagsStore();
const usersStore = useUsersStore();
const emit = defineEmits(["update"]);
const showDialog = ref(false);

const active = ref(true);
const action = ref<IFirewallRule["action"]>("allow");
const selectedIPOption = ref("all");
const selectedUsernameOption = ref("all");
const selectedFilterOption = ref(FormFilterOptions.All);

const {
  value: priority,
  errorMessage: priorityError,
  resetField: resetPriority,
} = useField<number>(
  "priority",
  yup.number()
    .integer()
    .required("This field is required")
    .notOneOf([0], "Priority cannot be zero")
    .typeError("This must be a valid integer"),
  { initialValue: 1 },
);

const {
  value: sourceIp,
  errorMessage: sourceIpError,
  setErrors: setSourceIpError,
  resetField: resetSourceIp,
} = useField<string>("sourceIp", yup.string().required("This field is required"), {
  initialValue: "",
});

const {
  value: username,
  errorMessage: usernameError,
  setErrors: setUsernameError,
  resetField: resetUsername,
} = useField<string>("username", yup.string().required("This field is required"), {
  initialValue: "",
});

const {
  value: hostname,
  errorMessage: hostnameError,
  setErrors: setHostnameError,
  resetField: resetHostname,
} = useField<string>("hostname", yup.string().required("This field is required"), {
  initialValue: "",
});

const selectedTags = ref<string[]>([]);
const selectedTagsError = ref("");

const activeSelectOptions = [
  { value: true, title: "Active" },
  { value: false, title: "Inactive" },
];
const actionSelectOptions = [
  { value: "allow", title: "Allow" },
  { value: "deny", title: "Deny" },
];
const sourceIPSelectOptions = [
  { value: "all", title: "Define source IP to all devices" },
  { value: "restrict", title: "Restrict source IP through a regexp" },
];
const usernameSelectOptions = [
  { value: "all", title: "Define rule to all users" },
  { value: "username", title: "Restrict access using a regexp for username" },
];
const filterSelectOptions = [
  { value: "all", title: "Define rule to all devices" },
  { value: "hostname", title: "Restrict rule with a regexp for hostname" },
  { value: "tags", title: "Restrict rule by device tags" },
];

const canCreateFirewallRule = hasPermission("firewall:create");

const acMenuOpen = ref(false);
const menuContentClass = computed(() => "fw-tags-ac-content");

const fetchedTags = ref<LocalTag[]>([]);
const tags = computed(() => fetchedTags.value);

const sentinel = ref<HTMLElement | null>(null);
let observer: IntersectionObserver | null = null;

const page = ref(1);
const perPage = ref(10);
const filter = ref("");
const isLoading = ref(false);

const hasMore = computed(() => tagsStore.numberTags > fetchedTags.value.length);

const setSelectedTagsError = () => {
  if (selectedFilterOption.value !== FormFilterOptions.Tags) {
    selectedTagsError.value = "";
    return;
  }
  if (selectedTags.value.length > 3) {
    nextTick(() => selectedTags.value.pop());
    selectedTagsError.value = "You can select up to 3 tags only.";
  } else if (selectedTags.value.length === 0) {
    selectedTagsError.value = "You must choose at least one tag";
  } else {
    selectedTagsError.value = "";
  }
};
watch(selectedTags, setSelectedTagsError);

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
  if (open && selectedFilterOption.value === FormFilterOptions.Tags) {
    await nextTick();
    setupObserver();
  } else {
    cleanupObserver();
  }
});

const resetSelectedTags = () => {
  selectedTags.value = [];
  selectedTagsError.value = "";
};

const handleFilterUpdate = async () => {
  resetHostname();
  resetSelectedTags();

  if (selectedFilterOption.value === FormFilterOptions.Hostname) setHostnameError("This field is required");
  if (selectedFilterOption.value === FormFilterOptions.Tags) {
    resetPagination();
    await loadTags();
    setSelectedTagsError();
  }
};

const handleSourceIpUpdate = () => {
  resetSourceIp();
  if (selectedIPOption.value === "restrict") setSourceIpError("This field is required");
};

const handleUsernameUpdate = () => {
  resetUsername();
  if (selectedUsernameOption.value === "username") setUsernameError("This field is required");
};

const hasErrors = computed(() => {
  const common = !!(
    priorityError.value
    || sourceIpError.value
    || usernameError.value
    || hostnameError.value
  );
  const tagsErrors = selectedFilterOption.value === FormFilterOptions.Tags
    && !!selectedTagsError.value;

  return common || tagsErrors;
});

const resetForm = () => {
  active.value = true;
  action.value = "allow";
  selectedFilterOption.value = FormFilterOptions.All;
  selectedIPOption.value = "all";
  selectedUsernameOption.value = "all";
  resetPriority();
  resetSourceIp();
  resetUsername();
  resetHostname();
  resetSelectedTags();
  cleanupObserver();
};

const open = () => {
  showDialog.value = true;
  if (selectedFilterOption.value === FormFilterOptions.Tags) {
    resetPagination();
    loadTags();
    setSelectedTagsError();
  }
};

const close = () => {
  showDialog.value = false;
  resetForm();
};

const update = () => {
  emit("update");
  close();
};

const constructNewFirewallRule = () => {
  const filterMap = {
    [FormFilterOptions.Hostname]: { hostname: hostname.value.trim() },
    [FormFilterOptions.Tags]: { tags: selectedTags.value },
    [FormFilterOptions.All]: { hostname: ".*" },
  };

  return {
    active: active.value,
    action: action.value,
    priority: Number(priority.value),
    source_ip: selectedIPOption.value === "all" ? ".*" : sourceIp.value.trim(),
    username: selectedUsernameOption.value === "all" ? ".*" : username.value.trim(),
    filter: filterMap[selectedFilterOption.value],
  };
};

const addFirewallRule = async () => {
  if (hasErrors.value) return;

  if (envVariables.isCommunity) {
    usersStore.showPaywall = true;
    return;
  }

  try {
    await firewallRulesStore.createFirewallRule(constructNewFirewallRule() as IFirewallRule);
    snackbar.showSuccess("Successfully created a new firewall rule.");
    update();
  } catch (error: unknown) {
    snackbar.showError("Failed to create a new firewall rule.");
    handleError(error);
  }
};

onUnmounted(() => {
  cleanupObserver();
});

defineExpose({ selectedIPOption, selectedUsernameOption, selectedFilterOption });
</script>
