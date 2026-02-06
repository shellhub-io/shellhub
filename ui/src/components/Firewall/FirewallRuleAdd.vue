<template>
  <div>
    <v-tooltip
      v-bind="$attrs"
      class="text-center"
      location="bottom"
      :disabled="canCreateFirewallRule"
    >
      <template #activator="{ props }">
        <div v-bind="props">
          <v-btn
            v-bind="$attrs"
            color="primary"
            tabindex="0"
            variant="elevated"
            :disabled="!canCreateFirewallRule"
            data-test="add-firewall-rule-btn"
            @click="open"
          >
            Add Rule
          </v-btn>
        </div>
      </template>
      <span> You don't have this kind of authorization. </span>
    </v-tooltip>

    <FormDialog
      v-model="showDialog"
      title="New Firewall Rule"
      icon="mdi-shield-check"
      confirm-text="Add"
      cancel-text="Cancel"
      :confirm-disabled="hasErrors"
      confirm-data-test="firewall-rule-add-btn"
      cancel-data-test="firewall-rule-cancel"
      data-test="firewall-rule-dialog"
      @close="close"
      @cancel="close"
      @confirm="addFirewallRule"
    >
      <v-card-text class="pa-6">
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

        <v-row class="mt-1 mb-3 px-3">
          <v-select
            v-model="selectedIPOption"
            label="Source IP access restriction"
            :items="sourceIPSelectOptions"
            hide-details
            data-test="firewall-rule-source-ip-select"
            @update:model-value="handleSourceIpUpdate"
          />
        </v-row>

        <v-text-field
          v-if="selectedIPOption === 'restrict'"
          v-model="sourceIp"
          label="Rule source IP"
          :error-messages="sourceIpError"
          hide-details="auto"
          data-test="firewall-rule-source-ip"
        />

        <v-row class="mt-5 mb-3 px-3">
          <v-select
            v-model="selectedUsernameOption"
            label="Device username access restriction"
            :items="usernameSelectOptions"
            hide-details
            data-test="username-field"
            @update:model-value="handleUsernameUpdate"
          />
        </v-row>

        <v-text-field
          v-if="selectedUsernameOption === 'username'"
          v-model="username"
          label="Username access restriction"
          placeholder="Username used during the connection"
          :error-messages="usernameError"
          hide-details="auto"
          data-test="firewall-rule-username-restriction"
        />

        <v-row class="mt-5 mb-3 px-3">
          <v-select
            v-model="selectedFilterOption"
            label="Device access restriction"
            :items="filterSelectOptions"
            hide-details
            data-test="filter-select"
            @update:model-value="handleFilterUpdate"
          />
        </v-row>

        <v-text-field
          v-if="selectedFilterOption === FormFilterOptions.Hostname"
          v-model="hostname"
          label="Device hostname access restriction"
          placeholder="Device hostname used during the connection"
          :error-messages="hostnameError"
          hide-details="auto"
          data-test="firewall-rule-hostname-restriction"
        />

        <v-row
          v-else-if="selectedFilterOption === FormFilterOptions.Tags"
          class="px-3 mt-3"
        >
          <TagAutocompleteSelect
            v-model:selected-tags="selectedTags"
            v-model:tag-selector-error-message="selectedTagsError"
          />
        </v-row>
      </v-card-text>
    </FormDialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
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
import useUsersStore from "@/store/modules/users";
import TagAutocompleteSelect from "@/components/Tags/TagAutocompleteSelect.vue";

const snackbar = useSnackbar();
const firewallRulesStore = useFirewallRulesStore();
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

const resetSelectedTags = () => {
  selectedTags.value = [];
  selectedTagsError.value = "";
};

const handleFilterUpdate = () => {
  resetHostname();
  resetSelectedTags();

  if (selectedFilterOption.value === FormFilterOptions.Hostname) setHostnameError("This field is required");
};

const handleSourceIpUpdate = () => {
  resetSourceIp();
  if (selectedIPOption.value === "restrict") setSourceIpError("This field is required");
};

const handleUsernameUpdate = () => {
  resetUsername();
  if (selectedUsernameOption.value === "username") setUsernameError("This field is required");
};

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

defineExpose({
  action,
  selectedIPOption, handleSourceIpUpdate,
  selectedUsernameOption, handleUsernameUpdate,
  selectedFilterOption, handleFilterUpdate,
  selectedTags, selectedTagsError,
});
</script>
