<template>
  <v-tooltip v-bind="$attrs" class="text-center" location="bottom" :disabled="hasAuthorization">
    <template v-slot:activator="{ props }">
      <div v-bind="props">
        <v-btn
          @click="dialog = !dialog"
          color="primary"
          tabindex="0"
          variant="elevated"
          aria-label="Dialog Add Public Key"
          :disabled="!hasAuthorization"
          @keypress.enter="dialog = !dialog"
          :size="size"
          data-test="public-key-add-btn"
        >
          Add Public Key
        </v-btn>
      </div>
    </template>
    <span> You don't have this kind of authorization. </span>
  </v-tooltip>

  <v-dialog v-model="dialog" width="520" transition="dialog-bottom-transition">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-3 bg-primary">
        New Public Key
      </v-card-title>
      <form @submit.prevent="create" class="mt-3">
        <v-card-text>
          <v-text-field
            v-model="name"
            :error-messages="nameError"
            label="Name"
            placeholder="Name used to identify the public key"
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
              data-test="access-restriction-field"
            />
          </v-row>

          <v-text-field
            v-if="choiceUsername === 'username'"
            v-model="username"
            label="Rule source IP"
            variant="underlined"
            :error-messages="usernameError"
          />

          <v-row class="mt-1 px-3">
            <v-select
              v-model="choiceFilter"
              label="Device access restriction"
              :items="filterList"
              variant="underlined"
              item-title="filterText"
              item-value="filterName"
              data-test="access-restriction-field"
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
import { useField } from "vee-validate";
import { computed, defineComponent, nextTick, ref, watch } from "vue";
import * as yup from "yup";
import { actions, authorizer } from "../../authorizer";
import { useStore } from "../../store";
import hasPermission from "../../utils/permission";
import { validateKey } from "../../utils/validate";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";

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
    const store = useStore();
    const dialog = ref(false);
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
    const supportedKeys = ref(
      "Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.",
    );

    const {
      value: name,
      errorMessage: nameError,
      setErrors: setnameError,
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
      if (role !== "") {
        return hasPermission(
          authorizer.role[role],
          actions.publicKey.create,
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
    watch(dialog, (value) => {
      if (!value) {
        setLocalVariable();
      }
    });

    const close = () => {
      dialog.value = false;
      setLocalVariable();
    };

    const update = () => {
      ctx.emit("update");
      close();
    };

    const hasErros = () => {
      if (name.value === "") {
        setnameError("This Field is required !");
        return true;
      }

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

    const resetFields = () => {
      resetName();
      resetUsername();
      resetHostname();
      resetPublicKeyData();
    };

    const create = async () => {
      if (!hasErros()) {
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
          store.dispatch(
            "snackbar/showSnackbarSuccessAction",
            INotificationsSuccess.publicKeyCreating,
          );
          update();
          resetFields();
        } catch (error: any) {
          if (error.response.status === 409) {
            setPublicKeyDataError("Public Key data already exists");
          } else {
            store.dispatch(
              "snackbar/showSnackbarErrorAction",
              INotificationsError.publicKeyCreating,
            );
            throw new Error(error);
          }
        }
      }
    };

    return {
      dialog,
      keyLocal,
      name,
      nameError,
      choiceUsername,
      usernameList,
      username,
      usernameError,
      choiceFilter,
      filterList,
      tagChoices,
      tagNames,
      validateLength,
      errMsg,
      hostname,
      hostnameError,
      publicKeyData,
      publicKeyDataError,
      supportedKeys,
      hasAuthorization,
      create,
      close,
    };
  },
});
</script>
