<template>
  <div>
    <v-tooltip v-bind="$attrs" class="text-center" location="bottom" :disabled="hasAuthorization">
      <template v-slot:activator="{ props }">
        <div v-bind="props">
          <v-btn
            v-bind="$attrs"
            @click="dialog = !dialog"
            color="primary"
            tabindex="0"
            variant="elevated"
            aria-label="Dialog Add device"
            :disabled="!hasAuthorization"
            @keypress.enter="dialog = !dialog"
            data-test="firewall-add-rule-btn"
          >
            Add Rule
          </v-btn>
        </div>
      </template>
      <span> You don't have this kind of authorization. </span>
    </v-tooltip>

    <v-dialog v-model="dialog" width="520" transition="dialog-bottom-transition">
      <v-card class="bg-v-theme-surface">
        <v-card-title class="text-h5 pa-3 bg-primary" data-test="firewall-rule-title">
          New Firewall Rule
        </v-card-title>
        <form @submit.prevent="create" class="mt-3">
          <v-card-text>
            <v-row>
              <v-col>
                <v-select
                  v-model="ruleFirewall.status"
                  :items="ruleStatus"
                  item-title="text"
                  item-value="type"
                  label="Rule status"
                  variant="underlined"
                  data-test="firewall-rule-status"
                />
              </v-col>

              <v-col>
                <v-text-field
                  v-model="ruleFirewall.priority"
                  label="Rule priority"
                  type="number"
                  variant="underlined"
                  :rules="[rules.required]"
                  data-test="firewall-rule-priority"
                />
              </v-col>

              <v-col>
                <v-select
                  v-model="ruleFirewall.policy"
                  :items="state"
                  item-title="name"
                  item-value="id"
                  label="Rule policy"
                  variant="underlined"
                  data-test="firewall-rule-policy"
                />
              </v-col>
            </v-row>

            <v-row class="mt-1 mb-1 px-3">
              <v-select
                v-model="choiceIP"
                label="Source IP access restriction"
                :items="sourceIPFieldChoices"
                item-title="filterText"
                item-value="filterName"
                variant="underlined"
                data-test="firewall-rule-source-ip"
              />
            </v-row>

            <v-text-field
              v-if="choiceIP === 'ipDetails'"
              v-model="sourceIp"
              label="Rule source IP"
              variant="underlined"
              :error-messages="sourceIpError"
              data-test="firewall-rule-source-ip-details"
            />

            <v-row class="mt-1 mb-1 px-3">
              <v-select
                v-model="choiceUsername"
                label="Device username access restriction"
                :items="usernameFieldChoices"
                item-title="filterText"
                item-value="filterName"
                variant="underlined"
                data-test="username-field"
              />
            </v-row>

            <v-text-field
              v-if="choiceUsername === 'username'"
              v-model="username"
              label="Username access restriction"
              placeholder="Username used during the connection"
              variant="underlined"
              :error-messages="usernameError"
              data-test="firewall-rule-username-restriction"
            />

            <v-row class="mt-2 mb-1 px-3">
              <v-select
                v-model="choiceFilter"
                label="Device access restriction"
                :items="filterFieldChoices"
                item-title="filterText"
                item-value="filterName"
                variant="underlined"
                data-test="device-field"
              />
            </v-row>

            <v-text-field
              v-if="choiceFilter === 'hostname'"
              v-model="filterField"
              label="Device hostname access restriction"
              placeholder="Device hostname used during the connection"
              :error-messages="filterFieldError"
              variant="underlined"
              data-test="firewall-rule-hostname-restriction"
            />

            <v-row v-if="choiceFilter === 'tags'" class="px-3 mt-2">
              <v-select
                v-model="tagChoices"
                :items="tagNames"
                data-test="tags-selector"
                attach
                chips
                label="Tags"
                :rules="[validateLength]"
                :error-messages="errMsg"
                variant="underlined"
                multiple
              />
            </v-row>
          </v-card-text>

          <v-card-actions>
            <v-spacer />
            <v-btn
              color="primary"
              @click="close"
              data-test="firewall-rule-cancel"
            >
              Cancel
            </v-btn>
            <v-btn
              color="primary"
              type="submit"
              data-test="firewall-rule-save-btn"
            >
              Save
            </v-btn>
          </v-card-actions>
        </form>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import { actions, authorizer } from "@/authorizer";
import hasPermission from "@/utils/permission";
import { envVariables } from "@/envVariables";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import { filterType } from "@/interfaces/IFirewallRule";
import useSnackbar from "@/helpers/snackbar";

export interface FirewallRuleType {
  action?: string;
  active?: boolean;
  policy?: string;
  priority?: number;
  status?: string;
  source_ip?: string;
  username?: string;
  filter?: filterType;
}

const snackbar = useSnackbar();
const store = useStore();
const emit = defineEmits(["update"]);
const dialog = ref(false);
const action = ref("create");
const choiceUsername = ref("all");
const choiceIP = ref("all");
const choiceFilter = ref("all");
const validateLength = ref(true);

const ruleFirewall = ref<FirewallRuleType>({
  policy: "allow",
  priority: 0,
  status: "active",
  source_ip: "",
  username: "",
});

const {
  value: sourceIp,
  errorMessage: sourceIpError,
  setErrors: setSourceIpError,
  resetField: resetSourceIp,
} = useField<string | undefined>("sourceIp", yup.string().required(), {
  initialValue: ruleFirewall.value.source_ip,
});

const {
  value: username,
  errorMessage: usernameError,
  setErrors: setUsernameError,
  resetField: resetUsername,
} = useField<string | undefined>("username", yup.string().required(), {
  initialValue: ruleFirewall.value.username,
});

const {
  value: filterField,
  errorMessage: filterFieldError,
  setErrors: setFilterFieldError,
  resetField: resetFilterField,
} = useField<string | undefined>("filterField", yup.string().required(), {
  initialValue: "",
});

watch(sourceIp, () => {
  ruleFirewall.value.source_ip = sourceIp.value;
});

watch(username, () => {
  ruleFirewall.value.username = username.value;
});

const errMsg = ref("");

const ruleStatus = ref([
  {
    type: "active",
    text: "Active",
  },
  {
    type: "inactive",
    text: "Inactive",
  },
]);

const tagChoices = ref([]);

const sourceIPFieldChoices = ref([
  {
    filterName: "all",
    filterText: "Define source IP to all devices",
  },
  {
    filterName: "ipDetails",
    filterText: "Restrict source IP through a regexp",
  },
]);

const filterFieldChoices = ref([
  {
    filterName: "all",
    filterText: "Define rule to all devices",
  },
  {
    filterName: "hostname",
    filterText: "Restrict rule with a regexp for hostname",
  },
  {
    filterName: "tags",
    filterText: "Restrict rule by device tags",
  },
]);

const usernameFieldChoices = ref([
  {
    filterName: "all",
    filterText: "Define rule to all users",
  },
  {
    filterName: "username",
    filterText: "Restrict access using a regexp for username",
  },
]);

const state = ref([
  {
    id: "allow",
    name: "Allow",
  },
  {
    id: "deny",
    name: "Deny",
  },
]);

const rules = ref({
  required: (value: string) => !!value || "Required.",
});

const tagNames = computed(() => store.getters["tags/list"]);

const hasAuthorization = computed(() => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(
      authorizer.role[role],
      actions.firewall[action.value],
    );
  }

  return false;
});

watch(tagChoices, (list) => {
  switch (true) {
    case list.length > 3:
      validateLength.value = false;
      nextTick(() => tagChoices.value.pop());
      errMsg.value = "The maximum capacity has reached";
      break;
    case list.length === 0:
      validateLength.value = false;
      errMsg.value = "You must choose at least one tag";
      break;
    default:
      validateLength.value = true;
      errMsg.value = "";
      break;
  }
});

const resetRuleFirewall = () => {
  ruleFirewall.value = {
    policy: "allow",
    priority: 0,
    status: "active",
    source_ip: "",
    username: "",
  };
  choiceFilter.value = "all";
  choiceIP.value = "all";
  choiceUsername.value = "all";
  tagChoices.value = [];
  validateLength.value = true;
  errMsg.value = "";
  resetSourceIp();
  resetUsername();
  resetFilterField();
};

const constructFilterObject = () => {
  let filterObj = {};

  switch (choiceFilter.value) {
    case "hostname":
      filterObj = { hostname: filterField };
      break;
    case "tags":
      filterObj = { tags: tagChoices };
      break;
    case "all":
      filterObj = { hostname: ".*" };
      break;
    default:
      break;
  }

  if (choiceUsername.value === "all") {
    ruleFirewall.value.username = ".*";
  }

  if (choiceIP.value === "all") {
    ruleFirewall.value.source_ip = ".*";
  }

  ruleFirewall.value = {
    ...ruleFirewall.value,
    filter: filterObj,
  };
};

const close = () => {
  dialog.value = false;
  resetRuleFirewall();
};

watch(choiceFilter, async () => {
  if (choiceFilter.value === "tags") {
    await store.dispatch("tags/fetch");
  }
});

const update = () => {
  emit("update");
  close();
};

const hasErrors = () => {
  if (
    choiceIP.value === "ipDetails"
        && ruleFirewall.value.source_ip === ""
  ) {
    setSourceIpError("This Field is required !");
    return true;
  }

  if (
    choiceUsername.value === "username"
        && ruleFirewall.value.username === ""
  ) {
    setUsernameError("This Field is required !");
    return true;
  }

  if (choiceFilter.value === "hostname" && filterField.value === "") {
    setFilterFieldError("This Field is required !");
    return true;
  }

  if (choiceFilter.value === "tags" && tagChoices.value.length === 0) {
    errMsg.value = "You must choose at least one tag";
    return true;
  }

  return false;
};

const create = async () => {
  if (!hasErrors()) {
    if (envVariables.isCommunity) {
      store.commit("users/setShowPaywall", true);
      return;
    }
    constructFilterObject();
    try {
      await store.dispatch("firewallRules/post", ruleFirewall.value);
      snackbar.showSuccess("Successfully created a new firewall rule.");
      update();
    } catch (error: unknown) {
      snackbar.showError("Failed to create a new firewall rule.");
      handleError(error);
    }
  }
};

defineExpose({ choiceIP, choiceFilter, choiceUsername });
</script>
