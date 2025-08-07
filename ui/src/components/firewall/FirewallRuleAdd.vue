<template>
  <div>
    <v-tooltip v-bind="$attrs" class="text-center" location="bottom" :disabled="hasAuthorization">
      <template v-slot:activator="{ props }">
        <div v-bind="props">
          <v-btn
            v-bind="$attrs"
            @click="open"
            color="primary"
            tabindex="0"
            variant="elevated"
            :disabled="!hasAuthorization"
            data-test="firewall-add-rule-btn"
          >
            Add Rule
          </v-btn>
        </div>
      </template>
      <span> You don't have this kind of authorization. </span>
    </v-tooltip>

    <BaseDialog v-model="showDialog" transition="dialog-bottom-transition">
      <v-card class="bg-v-theme-surface">
        <v-card-title class="text-h5 pa-3 bg-primary" data-test="firewall-rule-title">
          New Firewall Rule
        </v-card-title>
        <form @submit.prevent="addFirewallRule" class="mt-3">
          <v-card-text>
            <v-row>
              <v-col>
                <v-select
                  v-model="active"
                  :items="activeSelectOptions"
                  label="Rule status"
                  variant="underlined"
                  data-test="firewall-rule-status"
                />
              </v-col>

              <v-col>
                <v-text-field
                  v-model="priority"
                  label="Rule priority"
                  :error-messages="priorityError"
                  type="number"
                  variant="underlined"
                  data-test="firewall-rule-priority"
                />
              </v-col>

              <v-col>
                <v-select
                  v-model="action"
                  :items="actionSelectOptions"
                  label="Rule policy"
                  variant="underlined"
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
                variant="underlined"
                data-test="firewall-rule-source-ip-select"
              />
            </v-row>

            <v-text-field
              v-if="selectedIPOption === 'restrict'"
              v-model="sourceIp"
              label="Rule source IP"
              variant="underlined"
              :error-messages="sourceIpError"
              data-test="firewall-rule-source-ip"
            />

            <v-row class="mt-1 mb-1 px-3">
              <v-select
                v-model="selectedUsernameOption"
                @update:model-value="handleUsernameUpdate"
                label="Device username access restriction"
                :items="usernameSelectOptions"
                variant="underlined"
                data-test="username-field"
              />
            </v-row>

            <v-text-field
              v-if="selectedUsernameOption === 'username'"
              v-model="username"
              label="Username access restriction"
              placeholder="Username used during the connection"
              variant="underlined"
              :error-messages="usernameError"
              data-test="firewall-rule-username-restriction"
            />

            <v-row class="mt-2 mb-1 px-3">
              <v-select
                v-model="selectedFilterOption"
                @update:model-value="handleFilterUpdate"
                label="Device access restriction"
                :items="filterSelectOptions"
                variant="underlined"
                data-test="filter-select"
              />
            </v-row>

            <v-text-field
              v-if="selectedFilterOption === FormFilterOptions.Hostname"
              v-model="hostname"
              label="Device hostname access restriction"
              placeholder="Device hostname used during the connection"
              :error-messages="hostnameError"
              variant="underlined"
              data-test="firewall-rule-hostname-restriction"
            />

            <v-row v-else-if="selectedFilterOption === FormFilterOptions.Tags" class="px-3 mt-2">
              <v-select
                v-model="selectedTags"
                @update:model-value="setSelectedTagsError"
                :items="availableTags"
                data-test="tags-selector"
                attach
                chips
                label="Tags"
                :error-messages="selectedTagsError"
                variant="underlined"
                multiple
              />
            </v-row>
          </v-card-text>

          <v-card-actions>
            <v-spacer />
            <v-btn
              @click="close"
              data-test="firewall-rule-cancel"
            >
              Cancel
            </v-btn>
            <v-btn
              :disabled="hasErrors"
              color="primary"
              type="submit"
              data-test="firewall-rule-add-btn"
            >
              Add
            </v-btn>
          </v-card-actions>
        </form>
      </v-card>
    </BaseDialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import { IFirewallRule } from "@/interfaces/IFirewallRule";
import { actions, authorizer } from "@/authorizer";
import hasPermission from "@/utils/permission";
import { envVariables } from "@/envVariables";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import { FormFilterOptions } from "@/interfaces/IFilter";
import BaseDialog from "../BaseDialog.vue";

const snackbar = useSnackbar();
const store = useStore();
const emit = defineEmits(["update"]);
const showDialog = ref(false);
const active = ref(true);
const action = ref<IFirewallRule["action"]>("allow");
const selectedIPOption = ref("all");
const selectedUsernameOption = ref("all");
const selectedFilterOption = ref(FormFilterOptions.All);
const availableTags = computed(() => store.getters["tags/list"]);
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
  {
    initialValue: 1,
  },
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
const selectedTags = ref([]);
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

const hasAuthorization = computed(() => {
  const role = store.getters["auth/role"];
  return !!role && hasPermission(authorizer.role[role], actions.firewall.create);
});

const setSelectedTagsError = () => {
  if (selectedTags.value.length > 3) selectedTagsError.value = "You can select up to 3 tags only.";
  else if (selectedTags.value.length === 0) selectedTagsError.value = "You must choose at least one tag";
  else selectedTagsError.value = "";
};

const resetSelectedTags = () => {
  selectedTags.value = [];
  selectedTagsError.value = "";
};

const handleSourceIpUpdate = () => {
  resetSourceIp();
  if (selectedIPOption.value === "restrict") setSourceIpError("This field is required");
};

const handleUsernameUpdate = () => {
  resetUsername();
  if (selectedUsernameOption.value === "username") setUsernameError("This field is required");
};

const handleFilterUpdate = async () => {
  resetHostname();
  resetSelectedTags();

  if (selectedFilterOption.value === FormFilterOptions.Hostname) setHostnameError("This field is required");
  if (selectedFilterOption.value === FormFilterOptions.Tags) {
    setSelectedTagsError();
    await store.dispatch("tags/fetch");
  }
};

const hasErrors = computed(() => (
  !!(priorityError.value
    || sourceIpError.value
    || usernameError.value
    || hostnameError.value
    || selectedTagsError.value)
));

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
};

const open = () => {
  showDialog.value = true;
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
  const filter = {
    [FormFilterOptions.Hostname]: { hostname: hostname.value.trim() },
    [FormFilterOptions.Tags]: { tags: selectedTags.value },
    [FormFilterOptions.All]: { hostname: ".*" },
  }[selectedFilterOption.value];

  return {
    active: active.value,
    action: action.value,
    priority: Number(priority.value),
    source_ip: selectedIPOption.value === "all" ? ".*" : sourceIp.value.trim(),
    username: selectedUsernameOption.value === "all" ? ".*" : username.value.trim(),
    filter,
  };
};

const addFirewallRule = async () => {
  if (hasErrors.value) return;

  if (envVariables.isCommunity) {
    store.commit("users/setShowPaywall", true);
    return;
  }

  try {
    await store.dispatch("firewallRules/post", constructNewFirewallRule());
    snackbar.showSuccess("Successfully created a new firewall rule.");
    update();
  } catch (error: unknown) {
    snackbar.showError("Failed to create a new firewall rule.");
    handleError(error);
  }
};

defineExpose({ selectedIPOption, selectedFilterOption, selectedUsernameOption });
</script>
