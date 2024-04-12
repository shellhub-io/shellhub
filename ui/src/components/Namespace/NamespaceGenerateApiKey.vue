<template>
  <div>
    <v-tooltip location="bottom" class="text-center" :disabled="hasAuthorization">
      <template v-slot:activator="{ props }">
        <div v-bind="props">
          <v-btn
            :disabled="!hasAuthorization"
            color="primary"
            @click="dialog = !dialog"
            data-test="namespace-generate-main-btn"
          >
            Generate key
          </v-btn>
        </div>
      </template>
      <span> You don't have this kind of authorization. </span>
    </v-tooltip>

    <v-dialog v-model="dialog" @click:outside="close" max-width="450">
      <v-card data-test="namespace-generate-dialog" class="bg-v-theme-surface">
        <v-card-title class="bg-primary">New Api Key</v-card-title>

        <v-card-text data-test="namespace-generate-title">
          Generate a key that is scoped to the repository and is appropriate for personal API usage via HTTPS.
        </v-card-text>

        <v-card-text>
          <v-text-field
            v-model="keyName"
            :error-messages="keyInputError"
            label="Key Name"
            prepend-icon="mdi-key-outline"
            required
            variant="underlined"
            data-test="key-name-text"
            messages="Provide a distinct name for this key,
            which might be visible to resource owners or individuals in possession of the key."
          />
        </v-card-text>

        <v-card-text class="mt-2">
          <v-row>
            <v-col>
              <v-select
                v-model="selectedDate"
                label="Expiration date"
                :items="itemsDate"
                :item-props="true"
                :hint="expirationHint"
                variant="outlined"
                return-object
                data-test="namespace-generate-date"
              />
            </v-col>
            <v-col>
              <v-select
                v-if="hasAuthorization"
                v-model="selectedRole"
                label="Key Role"
                :items="itemsRoles"
                :item-props="true"
                variant="outlined"
                return-object
                data-test="namespace-generate-role"
              />
            </v-col>
          </v-row>
        </v-card-text>

        <v-card-text v-if="successKey">
          <v-alert
            text="Make sure to copy your key now as you will not be able to see it again."
            type="success"
            class="mb-2"
            data-test="successKey-alert"
          />
          <v-text-field
            v-model="keyResponse"
            append-inner-icon="mdi-content-copy"
            variant="solo-filled"
            readonly
            density="compact"
            @click="copyText(keyResponse)"
            data-test="keyResponse-text"
          />
        </v-card-text>

        <v-card-text v-if="failKey">
          <v-alert
            :text="errorMessage"
            type="error"
            class="mb-2"
            data-test="failMessage-alert"
          />
        </v-card-text>

        <v-card-actions>
          <v-btn data-test="close-btn" @click="close()"> Close </v-btn>
          <v-spacer />

          <v-btn color="success" variant="flat" data-test="add-btn" @click="generateKey()" :disabled="successKey">
            Generate Api Key
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import moment from "moment";
import * as yup from "yup";
import axios, { AxiosError } from "axios";
import { useField } from "vee-validate";
import hasPermission from "../../utils/permission";
import { useStore } from "@/store";
import { actions, authorizer } from "@/authorizer";
import handleError from "@/utils/handleError";
import { INotificationsCopy, INotificationsError } from "@/interfaces/INotifications";

const emit = defineEmits(["update"]);

const store = useStore();

const hasAuthorization = computed(() => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(
      authorizer.role[role],
      actions.apiKey.create,
    );
  }
  return false;
});

const dialog = ref(false);
const successKey = ref(false);
const failKey = ref(false);
const errorMessage = ref("");
const keyResponse = computed(() => store.getters["apiKeys/apiKey"]);
const isOwner = computed(() => store.getters["auth/role"] === "owner");
const {
  value: keyName,
  errorMessage: keyInputError,
} = useField<string | undefined>(
  "name",
  yup
    .string()
    .required()
    .min(3)
    .max(20)
    .matches(/^(?!.*\s).*$/, "This field cannot contain any blankspaces"),
  {
    initialValue: "",
  },
);

const copyText = (value: string | undefined) => {
  if (value) {
    navigator.clipboard.writeText(value);
    store.dispatch(
      "snackbar/showSnackbarCopy",
      INotificationsCopy.copyKey,
    );
  }
};

const getExpiryDate = (item) => {
  if (item === "No expire") {
    return {
      expirationDate: "This key will never expire",
      expirationDateSelect: "Expires never",
    };
  }

  const [value, unit] = item.split(" ");
  const expirationDate = `This key expires in ${moment().add(value, unit).format("MMMM, YYYY")}`;
  const expirationDateSelect = `Expires in ${moment().add(value, unit).format("MMMM, YYYY")}`;
  return {
    expirationDate,
    expirationDateSelect,
  };
};

const itemsRoles = [
  {
    title: "observer",
    value: "observer",
  },
  {
    title: "operator",
    value: "operator",
  },
  {
    title: "administrator",
    value: "administrator",
    disabled: !hasAuthorization.value || !isOwner.value,
  },
];

const itemsDate = [
  {
    title: "30 days",
    subtitle: getExpiryDate("30 days").expirationDateSelect,
    time: 30,
  },
  {
    title: "60 days",
    subtitle: getExpiryDate("60 days").expirationDateSelect,
    time: 60,
  },
  {
    title: "90 days",
    subtitle: getExpiryDate("90 days").expirationDateSelect,
    time: 90,
  },
  {
    title: "1 year",
    subtitle: getExpiryDate("1 year").expirationDateSelect,
    time: 365,
  },
  {
    title: "No expire",
    subtitle: getExpiryDate("No expire").expirationDateSelect,
    time: -1,
  },
];

const selectedDate = ref(itemsDate[0]);
const selectedRole = ref(itemsRoles[0]);
const expirationHint = ref(getExpiryDate(selectedDate.value.title).expirationDate);
const tenant = computed(() => localStorage.getItem("tenant"));

watch(selectedDate, (newVal) => {
  expirationHint.value = getExpiryDate(newVal.title).expirationDate;
});

const generateKey = async () => {
  try {
    await store.dispatch("apiKeys/generateApiKey", {
      tenant: tenant.value,
      name: keyName.value,
      expires_at: selectedDate.value.time,
      role: selectedRole.value.title,
    });
    successKey.value = true;
    failKey.value = false;
    emit("update");
  } catch (error: unknown) {
    failKey.value = true;
    successKey.value = false;
    store.dispatch(
      "snackbar/showSnackbarErrorAction",
      INotificationsError.generateKey,
    );
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      switch (axiosError.response?.status) {
        case 400:
          errorMessage.value = "Please provide a name for the API key.";
          break;
        case 401:
          errorMessage.value = "You are not authorized to create an API key.";
          break;
        case 409:
          errorMessage.value = "An API key with the same name already exists.";
          break;
        default:
          errorMessage.value = "An error occurred while generating your API key. Please try again later.";
          handleError(error);
      }
      return;
    }
    handleError(error);
  }
};

const close = () => {
  failKey.value = false;
  dialog.value = false;
  successKey.value = false;
  keyName.value = "";
  [selectedDate.value] = itemsDate;
  [selectedRole.value] = itemsRoles;
};
defineExpose({ errorMessage });
</script>
