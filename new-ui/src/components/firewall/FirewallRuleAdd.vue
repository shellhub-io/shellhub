<template>
  <v-btn
    v-bind="$attrs"
    @click="dialog = !dialog"
    color="primary"
    tabindex="0"
    variant="elevated"
    aria-label="Dialog Add device"
    @keypress.enter="dialog = !dialog"
    :size="size"
    data-test="device-add-btn"
  >
    Add Rule
  </v-btn>

  <v-dialog v-model="dialog" width="520" transition="dialog-bottom-transition">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-3 bg-primary">
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
              />
            </v-col>

            <v-col>
              <v-text-field
                v-model="ruleFirewall.priority"
                label="Rule priority"
                type="number"
                variant="underlined"
                :rules="[rules.required]"
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
              data-test="source_ip-field"
            />
          </v-row>

          <v-text-field
            v-if="choiceIP === 'ipDetails'"
            v-model="sourceIp"
            label="Rule source IP"
            variant="underlined"
            :error-messages="sourceIpError"
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
            data-test="hostname-field"
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
            text
            @click="close"
            data-test="device-add-cancel-btn"
          >
            Cancel
          </v-btn>
          <v-btn
            color="primary"
            text
            type="submit"
            data-test="device-add-save-btn"
          >
            Save
          </v-btn>
        </v-card-actions>
      </form>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { computed, defineComponent, nextTick, ref, watch } from "vue";
import { actions, authorizer } from "../../authorizer";
import hasPermission from "../../utils/permission";
import { useStore } from "../../store";
import { useField } from "vee-validate";
import * as yup from "yup";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";

export interface FirewallRuleType {
  policy?: string;
  priority?: number;
  status?: string;
  source_ip?: string;
  username?: string;
  filter?: any;
}

export default defineComponent({
  props: {
    size: {
      type: String,
      default: "default",
      required: false,
    },
  },
  emits: ["update"],
  setup(props, ctx) {
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

    const store = useStore();

    const tagNames = computed(() => store.getters["tags/list"]);

    const hasAuthorization = computed(() => {
      const role = store.getters["auth/role"];
      if (role !== "") {
        return hasPermission(
          authorizer.role[role],
          actions.firewall[action.value]
        );
      }

      return false;
    });

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

      if (choiceFilter.value === "hostname") {
        filterObj = { hostname: filterField };
      } else if (choiceFilter.value === "tags") {
        filterObj = { tags: tagChoices };
      } else if (choiceFilter.value === "all") {
        filterObj = { hostname: ".*" };
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

    const update = () => {
      ctx.emit("update");
      close();
    };

    const close = () => {
      dialog.value = false;
      resetRuleFirewall();
    };

    const hasErros = () => {
      if (
        choiceIP.value === "ipDetails" &&
        ruleFirewall.value.source_ip === ""
      ) {
        setSourceIpError("This Field is required !");
        return true;
      }

      if (
        choiceUsername.value === "username" &&
        ruleFirewall.value.username === ""
      ) {
        setUsernameError("This Field is required !");
        return true;
      }

      if (choiceFilter.value === "hostname" && filterField.value === "") {
        setFilterFieldError("This Field is required !");
        return true;
      }

      return false;
    };

    const create = async () => {
      if (!hasErros()) {
        constructFilterObject();
        try {
          await store.dispatch("firewallRules/post", ruleFirewall.value);
          store.dispatch(
            "snackbar/showSnackbarSuccessAction",
            INotificationsSuccess.firewallRuleCreating
          );
          update();
        } catch (error: any) {
          store.dispatch(
            "snackbar/showSnackbarErrorAction",
            INotificationsError.firewallRuleCreating
          );
          throw new Error(error);
        }
      }
    };

    return {
      dialog,
      action,
      choiceUsername,
      choiceIP,
      choiceFilter,
      validateLength,
      filterField,
      filterFieldError,
      sourceIp,
      sourceIpError,
      username,
      usernameError,
      errMsg,
      ruleStatus,
      tagChoices,
      sourceIPFieldChoices,
      filterFieldChoices,
      usernameFieldChoices,
      ruleFirewall,
      state,
      tagNames,
      hasAuthorization,
      rules,
      create,
      close,
      constructFilterObject,
    };
  },
});
</script>
