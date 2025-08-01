<template>
  <div>
    <v-tooltip v-bind="$attrs" class="text-center" location="bottom" :disabled="hasAuthorization">
      <template v-slot:activator="{ props }">
        <div v-bind="props">
          <v-btn
            @click="showDialog = true"
            color="primary"
            tabindex="0"
            variant="elevated"
            aria-label="Add Public Key"
            :disabled="!hasAuthorization"
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

    <BaseDialog v-model="showDialog" transition="dialog-bottom-transition">
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
              <v-select
                v-if="choiceFilter === 'tags'"
                v-model="tagChoices"
                :items="tagNames"
                data-test="tags-selector"
                attach
                chips
                label="Tags"
                :rules="[validateLength]"
                :error-messages="errMsg"
                multiple
              />
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
              data-test="pk-add-cancel-btn"
            >
              Cancel
            </v-btn>
            <v-btn
              color="primary"
              type="submit"
              data-test="pk-add-save-btn"
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
import { computed, nextTick, ref, watch } from "vue";
import * as yup from "yup";
import axios, { AxiosError } from "axios";
import { actions, authorizer } from "@/authorizer";
import { useStore } from "@/store";
import hasPermission from "@/utils/permission";
import { validateKey } from "@/utils/validate";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";

const { size } = defineProps<{ size?: string }>();

const emit = defineEmits(["update"]);
const store = useStore();
const showDialog = ref(false);
const snackbar = useSnackbar();
const validateLength = ref(true);
const choiceFilter = ref("all");
const choiceUsername = ref("all");
const tagChoices = ref([]);
const errMsg = ref("");
const keyLocal = ref({});
const usernameList = ref([
  {
    filterName: "all",
    filterText: "Allow any user",
  },
  {
    filterName: "username",
    filterText: "Restrict access using a regexp for username",
  },
]);
const filterList = ref([
  {
    filterName: "all",
    filterText: "Allow the key to connect to all available devices",
  },
  {
    filterName: "hostname",
    filterText: "Restrict access using a regexp for hostname",
  },
  {
    filterName: "tags",
    filterText: "Restrict access by tags",
  },
]);

const {
  value: name,
  errorMessage: nameError,
  resetField: resetName,
} = useField<string | undefined>("name", yup.string().required(), {
  initialValue: "",
});

const {
  value: username,
  errorMessage: usernameError,
  setErrors: setUsernameError,
  resetField: resetUsername,
} = useField<string | undefined>("username", yup.string().required(), {
  initialValue: "",
});

const {
  value: hostname,
  errorMessage: hostnameError,
  setErrors: setHostnameError,
  resetField: resetHostname,
} = useField<string | undefined>("hostname", yup.string().required(), {
  initialValue: "",
});

const {
  value: publicKeyData,
  errorMessage: publicKeyDataError,
  setErrors: setPublicKeyDataError,
  resetField: resetPublicKeyData,
} = useField<string>("publicKeyData", yup.string().required(), {
  initialValue: "",
});

const tagNames = computed(() => store.getters["tags/list"]);

const hasAuthorization = computed(() => {
  const role = store.getters["auth/role"];
  return !!role && hasPermission(authorizer.role[role], actions.publicKey.create);
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

watch(choiceFilter, async () => {
  if (choiceFilter.value === "tags") {
    await store.dispatch("tags/fetch");
  }
});

watch(publicKeyData, async () => {
  if (publicKeyData.value !== "") {
    setPublicKeyDataError("Field is required");
  }

  if (await validateKey("public", publicKeyData.value)) {
    setPublicKeyDataError("This is not valid key");
  }
});

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
    default:
  }
};

const chooseFilter = () => {
  switch (choiceFilter.value) {
    case "all": {
      keyLocal.value = { ...keyLocal.value, filter: { hostname: ".*" } };
      break;
    }
    case "hostname": {
      keyLocal.value = {
        ...keyLocal.value,
        filter: { hostname: hostname.value },
      };
      break;
    }
    case "tags": {
      keyLocal.value = {
        ...keyLocal.value,
        filter: { tags: tagChoices.value },
      };
      break;
    }
    default:
  }
};

const setLocalVariable = () => {
  keyLocal.value = {};
  hostname.value = "";
  tagChoices.value = [];
  choiceFilter.value = "all";
  choiceUsername.value = "all";
};

watch(showDialog, (value) => {
  if (!value) {
    setLocalVariable();
  }
});

const close = () => {
  showDialog.value = false;
  setLocalVariable();
};

const update = () => {
  emit("update");
  close();
};

const hasErrors = () => {
  if (choiceUsername.value === "username" && username.value === "") {
    setUsernameError("This Field is required!");
    return true;
  }

  if (choiceFilter.value === "hostname" && hostname.value === "") {
    setHostnameError("This Field is required!");
    return true;
  }

  if (choiceFilter.value === "tags" && tagChoices.value.length === 0) {
    return true;
  }

  return false;
};

const resetFields = () => {
  resetName();
  resetUsername();
  resetHostname();
  resetPublicKeyData();
};

const create = async () => {
  if (!hasErrors()) {
    try {
      chooseFilter();
      chooseUsername();
      const keySend = {
        ...keyLocal.value,
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        data: btoa(publicKeyData.value),
        name: name.value,
      };
      await store.dispatch("publicKeys/post", keySend);
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
  }
};

defineExpose({ publicKeyDataError, nameError });
</script>
