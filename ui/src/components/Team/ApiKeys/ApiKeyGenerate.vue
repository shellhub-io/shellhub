<template>
  <div>
    <v-tooltip location="bottom" class="text-center" :disabled="hasAuthorization">
      <template v-slot:activator="{ props }">
        <div v-bind="props">
          <v-btn
            :disabled="!hasAuthorization"
            color="primary"
            @click="showDialog = true"
            data-test="api-key-generate-main-btn"
          >
            Generate key
          </v-btn>
        </div>
      </template>
      <span> You don't have this kind of authorization. </span>
    </v-tooltip>

    <BaseDialog v-model="showDialog" @update:modelValue="(value: boolean) => !value && close()">
      <v-card data-test="api-key-generate-dialog" class="bg-v-theme-surface">
        <v-card-title class="bg-primary">New Api Key</v-card-title>

        <v-card-text v-if="errorMessage">
          <v-alert
            :text="errorMessage"
            type="error"
            class="mt-1"
            data-test="fail-message-alert"
          />
        </v-card-text>

        <v-card-text data-test="api-key-generate-title">
          Generate a key that is scoped to the namespace and is appropriate for personal API usage via HTTPS.

          <v-text-field
            class="mt-6"
            v-model="keyName"
            :error-messages="keyNameError"
            label="Key Name"
            prepend-inner-icon="mdi-key-outline"
            required
            data-test="key-name-text"
            messages="Provide a distinct name for this key,
          which might be visible to resource owners or individuals in possession of the key."
          />
          <v-row class="mt-6">
            <v-col>
              <v-select
                v-model="selectedDate"
                label="Expiration date"
                :items="itemsDate"
                :item-props="true"
                :hint="expirationHint"
                return-object
                data-test="api-key-generate-date"
              />
            </v-col>
            <v-col>
              <RoleSelect
                v-if="hasAuthorization"
                v-model="selectedRole"
                data-test="api-key-generate-role"
              />
            </v-col>
          </v-row>

        </v-card-text>
        <v-card-text v-if="generatedApiKey">
          <v-alert
            text="Make sure to copy your key now as you will not be able to see it again."
            type="success"
            class="mb-2"
            data-test="success-key-alert"
          />
          <CopyWarning :copied-item="'API Key'">
            <template #default="{ copyText }">
              <v-text-field
                v-model="generatedApiKey"
                append-inner-icon="mdi-content-copy"
                variant="solo-filled"
                readonly
                density="compact"
                @click="copyText(generatedApiKey)"
                data-test="key-response-text"
              />
            </template>
          </CopyWarning>

        </v-card-text>

        <v-card-actions>
          <v-btn data-test="close-btn" @click="close()"> Close </v-btn>
          <v-spacer />

          <v-btn color="success" variant="flat" data-test="add-btn" @click="generateKey()" :disabled="!!generatedApiKey || !!keyNameError">
            Generate Api Key
          </v-btn>
        </v-card-actions>
      </v-card>
    </BaseDialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import moment from "moment";
import * as yup from "yup";
import axios from "axios";
import { useField } from "vee-validate";
import hasPermission from "@/utils/permission";
import { useStore } from "@/store";
import { actions, authorizer } from "@/authorizer";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import CopyWarning from "@/components/User/CopyWarning.vue";
import BaseDialog from "@/components/BaseDialog.vue";
import RoleSelect from "@/components/Team/RoleSelect.vue";
import { BasicRole } from "@/interfaces/INamespace";
import useApiKeysStore from "@/store/modules/api_keys";

const emit = defineEmits(["update"]);
const snackbar = useSnackbar();
const store = useStore();
const apiKeyStore = useApiKeysStore();
const showDialog = ref(false);
const errorMessage = ref("");
const generatedApiKey = ref("");
const hasAuthorization = computed(() => {
  const role = store.getters["auth/role"];
  return !!role && hasPermission(authorizer.role[role], actions.apiKey.create);
});

const {
  value: keyName,
  errorMessage: keyNameError,
} = useField<string>(
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

const getExpiryDate = (item) => {
  if (item === "No expire") {
    return {
      expirationDate: "This key will never expire",
      expirationDateSelect: "Never Expires",
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
const selectedRole = ref<BasicRole>("administrator");
const expirationHint = ref(getExpiryDate(selectedDate.value.title).expirationDate);

watch(selectedDate, (newVal) => {
  expirationHint.value = getExpiryDate(newVal.title).expirationDate;
});

const handleGenerateKeyError = (error: unknown) => {
  snackbar.showError("Failed to generate API Key.");

  if (axios.isAxiosError(error)) {
    const status = error.response?.status;
    switch (status) {
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
  } else {
    handleError(error);
  }
};

const generateKey = async () => {
  try {
    generatedApiKey.value = await apiKeyStore.generateApiKey({
      name: keyName.value,
      expires_in: selectedDate.value.time,
      role: selectedRole.value,
    });
    emit("update");
  } catch (error: unknown) {
    handleGenerateKeyError(error);
  }
};

const close = () => {
  showDialog.value = false;
  generatedApiKey.value = "";
  keyName.value = "";
  [selectedDate.value] = itemsDate;
  selectedRole.value = "administrator";
};
defineExpose({ errorMessage });
</script>
