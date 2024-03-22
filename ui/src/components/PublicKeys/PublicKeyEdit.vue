<template>
  <div>
    <v-list-item
      @click="open()"
      v-bind="$attrs"
      :disabled="notHasAuthorization"
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

    <v-dialog v-model="showDialog" width="520" transition="dialog-bottom-transition">
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
              label="Rule source IP"
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
                variant="underlined"
                multiple
              />
              <v-text-field
                v-if="choiceFilter === 'hostname'"
                v-model="hostname"
                label="Hostname"
                variant="underlined"
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
              :messages="supportedKeys"
              variant="underlined"
              data-test="data-field"
              rows="2"
            />
          </v-card-text>

          <v-card-actions>
            <v-spacer />
            <v-btn
              color="primary"
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
    </v-dialog>
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
} from "vue";
import * as yup from "yup";
import { useStore } from "../../store";
import { IPublicKey } from "../../interfaces/IPublicKey";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import handleError from "../../utils/handleError";

const props = defineProps({
  show: {
    type: Boolean,
    required: false,
  },
  keyObject: {
    type: Object,
    required: true,
    default: Object as unknown as IPublicKey,
  },
  notHasAuthorization: {
    type: Boolean,
    default: false,
  },
});
const emit = defineEmits(["update"]);
const showDialog = ref(false);
const store = useStore();
const choiceFilter = ref("hostname");
const validateLength = ref(true);
const errMsg = ref("");
const prop = computed(() => props);
const choiceUsername = ref("username");
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
const tagChoices = ref<Array<string>>([]);
const keyLocal = ref<Partial<IPublicKey>>({
  name: "",
  username: "",
  data: "",
});
const supportedKeys = ref(
  "Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.",
);

const {
  value: name,
  errorMessage: nameError,
} = useField<string>("name", yup.string().required(), {
  initialValue: prop.value.keyObject.name,
});

watch(name, () => {
  keyLocal.value.name = name.value;
});

const {
  value: username,
  errorMessage: usernameError,
  setErrors: setUsernameError,
} = useField<string>("username", yup.string().required(), {
  initialValue: prop.value.keyObject.username,
});

watch(username, () => {
  keyLocal.value.username = username.value;
});

const {
  value: hostname,
  errorMessage: hostnameError,
  setErrors: setHostnameError,
} = useField<string>("hostname", yup.string().required(), {
  initialValue: prop.value.keyObject.filter?.hostname || "",
});

const {
  value: publicKeyData,
  errorMessage: publicKeyDataError,
} = useField<string>("publicKeyData", yup.string().required(), {
  initialValue: prop.value.keyObject.data,
});

const hasTags = computed(() => {
  const { keyObject } = props;
  if (!keyObject) return false;
  return Reflect.ownKeys(keyObject.filter)[0] === "tags";
});

const tagNames = computed({
  get() {
    return store.getters["tags/list"];
  },
  set(val) {
    tagChoices.value = val;
  },
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
      const { tags } = props.keyObject.filter;
      tagChoices.value = tags;
      choiceFilter.value = "tags";
    } else {
      const { hostname: hostnameLocal } = props.keyObject.filter;
      if (!!hostnameLocal && hostnameLocal !== ".*") {
        choiceFilter.value = "hostname";
        hostname.value = hostnameLocal;
      } else if (!!hostnameLocal && hostnameLocal === ".*") {
        choiceFilter.value = "all";
      }
    }

    const { username: usernameLocal } = props.keyObject;
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

const setLocalVariable = () => {
  keyLocal.value = { ...props.keyObject };
  keyLocal.value.data = atob(props.keyObject.data);
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
  publicKeyData.value = props.keyObject.data;
};

onMounted(async () => {
  await setLocalVariable();
});

onUpdated(async () => {
  handleUpdate();
  await setLocalVariable();
  keyLocal.value.data = publicKeyData.value;
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
};

const update = () => {
  emit("update");
  close();
};

const edit = async () => {
  if (!hasError()) {
    chooseFilter();
    chooseUsername();
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    const keySend = { ...keyLocal.value, data: btoa(keyLocal.value.data) };

    try {
      await store.dispatch("publicKeys/put", keySend);
      store.dispatch(
        "snackbar/showSnackbarSuccessAction",
        INotificationsSuccess.publicKeyEditing,
      );
      update();
    } catch (error: unknown) {
      store.dispatch(
        "snackbar/showSnackbarErrorAction",
        INotificationsError.publicKeyEditing,
      );
      handleError(error);
    }
  }
};

defineExpose({ nameError, usernameError, hostnameError });
</script>
