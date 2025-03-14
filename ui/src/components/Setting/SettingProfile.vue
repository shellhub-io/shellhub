<template>
  <v-container
    fluid
    data-test="account-profile-container"
  >
    <UserDelete v-model:show="showDeleteAccountDialog" data-test="delete-user-dialog" />
    <v-card
      variant="flat"
      class="bg-transparent"
      data-test="account-profile-card"
    >
      <v-card-item>
        <v-list-item
          class="pa-0"
          data-test="profile-header"
        >
          <template v-slot:prepend>
            <UserIcon size="4rem" data-test="user-icon" />
          </template>
          <template v-slot:title>
            <h1 data-test="profile-title">Account Profile</h1>
          </template>
          <template v-slot:subtitle>
            <span data-test="profile-subtitle">Manage your account profile</span>
          </template>
          <template v-slot:append>
            <div class="mr-4">
              <v-btn
                v-if="!editDataStatus"
                @click="editDataStatus = !editDataStatus"
                color="primary"
                variant="text"
                class="bg-secondary border"
                data-test="edit-profile-button"
              >Edit Profile</v-btn>
              <template v-else>
                <v-btn
                  @click="cancel('data')"
                  color="primary"
                  variant="text"
                  class="mr-2"
                  data-test="cancel-edit-button"
                >Cancel</v-btn>
                <v-btn
                  @click="updateUserData"
                  color="primary"
                  variant="flat"
                  data-test="save-changes-button"
                >Save Changes</v-btn>
              </template>
            </div>
          </template>
        </v-list-item>
      </v-card-item>

      <v-card-text class="pt-4">
        <v-list
          border
          rounded
          class="bg-background pa-0"
          data-test="profile-details-list"
        >
          <v-card-item style="grid-template-columns: max-content 1.5fr 2fr">
            <template #prepend>
              <v-icon>mdi-badge-account</v-icon>
            </template>
            <template #title>
              <span class="text-subtitle-1" data-test="name-field">Name</span>
            </template>
            <template #append>
              <v-text-field
                v-model="name"
                :error-messages="nameError"
                :disabled="!editDataStatus"
                :readonly="!editDataStatus"
                required
                :hide-details="!nameError"
                density="compact"
                :variant="editDataStatus ? 'outlined' : 'plain'"
                data-test="name-input"
              />
            </template>
          </v-card-item>
          <v-divider />
          <div v-if="isLocalAuth || isCloud">
            <v-card-item style="grid-template-columns: max-content 1.5fr 2fr">
              <template #prepend>
                <v-icon>mdi-account</v-icon>
              </template>
              <template #title>
                <span class="text-subtitle-1" data-test="username-field">Username</span>
              </template>
              <template #append>
                <v-text-field
                  v-model="username"
                  :error-messages="usernameError"
                  :disabled="!editDataStatus"
                  :readonly="!editDataStatus"
                  density="compact"
                  :variant="editDataStatus ? 'outlined' : 'plain'"
                  required
                  :hide-details="!usernameError"
                  data-test="username-input"
                />
              </template>
            </v-card-item>
            <v-divider />
          </div>
          <v-card-item style="grid-template-columns: max-content 1.5fr 2fr">
            <template #prepend>
              <v-icon>mdi-email</v-icon>
            </template>
            <template #title>
              <span class="text-subtitle-1" data-test="email-field">Email</span>
            </template>
            <template #append>
              <v-text-field
                v-model="email"
                :error-messages="emailError"
                :disabled="!editDataStatus"
                :readonly="!editDataStatus"
                density="compact"
                :variant="editDataStatus ? 'outlined' : 'plain'"
                required
                :hide-details="!emailError"
                data-test="email-input"
              />
            </template>
          </v-card-item>
          <v-divider />

          <div v-if="isLocalAuth || isCloud">
            <v-card-item style="grid-template-columns: max-content 1.5fr 2fr">
              <template #prepend>
                <v-icon>mdi-email-lock</v-icon>
              </template>
              <template #title>
                <span class="text-subtitle-1" data-test="recovery-email-field">Recovery Email</span>
              </template>
              <template #append>
                <v-text-field
                  v-model="recoveryEmail"
                  :error-messages="recoveryEmailError"
                  :disabled="!editDataStatus"
                  :readonly="!editDataStatus"
                  density="compact"
                  :variant="editDataStatus ? 'outlined' : 'plain'"
                  required
                  :hide-details="!recoveryEmailError"
                  data-test="recovery-email-input"
                />
              </template>
            </v-card-item>
            <v-divider />
            <v-card-item style="grid-template-columns: max-content 1.5fr 2fr" v-if="isLocalAuth || isCloud">
              <template #prepend>
                <v-icon>mdi-key</v-icon>
              </template>
              <template #title>
                <span class="text-subtitle-1">Password</span>
              </template>
              <template #append>
                <v-btn variant="text" color="aprimary" @click="showChangePassword = true">Change Password</v-btn>
                <ChangePassword v-model="showChangePassword" />
              </template>
            </v-card-item>
            <v-divider />
            <div class="d-flex align-center justify-space-between pr-4">
              <v-card
                :disabled="isCommunity"
                flat
                class="bg-background"
                :class="lgAndUp ? 'w-100' : 'w-75'"
                prepend-icon="mdi-fingerprint"
                data-test="mfa-card"
              >
                <template #title>
                  <span class="text-subtitle-1">Multi-factor Authentication</span>
                </template>
                <div class="d-flex flex-no-wrap justify-space-between">
                  <div>
                    <v-card-text class="pt-0 text-justify" data-test="mfa-text">
                      Enable multi-factor authentication (MFA) to add an extra layer of security to your account.
                      You'll need to enter a one-time verification code from your preferred TOTP provider to log in.
                    </v-card-text>
                  </div>
                </div>
              </v-card>
              <v-tooltip location="top" text="Only available for Cloud or Enterprise accounts!" :disabled="!isCommunity">
                <template v-slot:activator="{ props }">
                  <div v-bind="props" class="d-flex align-center bg-background" style="height: fit-content;">
                    <v-switch
                      hide-details
                      inset
                      color="primary"
                      v-model="mfaEnabled"
                      @click="toggleMfa()"
                      :disabled="isCommunity"
                      data-test="switch-mfa"
                    />
                    <MfaSettings v-model="dialogMfaSettings" />
                    <MfaDisable v-model="dialogMfaDisable" />
                  </div>
                </template>
              </v-tooltip>
            </div>

          </div>
          <v-divider />
          <v-card-item style="grid-template-columns: max-content 1.5fr 2fr">
            <template #prepend>
              <v-icon>mdi-delete</v-icon>
            </template>
            <template #title>
              <span class="text-subtitle-1" data-test="delete-account">Delete Account</span>
            </template>
            <template #append>
              <v-btn variant="text" color="error" @click="showDeleteAccountDialog = true" data-test="delete-account-btn">Delete</v-btn>
            </template>
          </v-card-item>
        </v-list>
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
/* eslint-disable @typescript-eslint/ban-ts-comment */
import { ref, computed, onMounted } from "vue";
import { useDisplay } from "vuetify";
import { useField } from "vee-validate";
import axios, { AxiosError } from "axios";
import * as yup from "yup";
import { useStore } from "@/store";
import { INotificationsSuccess } from "@/interfaces/INotifications";
import handleError from "@/utils/handleError";
import MfaSettings from "../AuthMFA/MfaSettings.vue";
import MfaDisable from "../AuthMFA/MfaDisable.vue";
import UserDelete from "../User/UserDelete.vue";
import UserIcon from "../User/UserIcon.vue";
import { envVariables } from "@/envVariables";
import ChangePassword from "../User/ChangePassword.vue";

type ErrorResponseData = { field: string; message: string }[];

const store = useStore();
const editDataStatus = ref(false);
const editPasswordStatus = ref(false);
const mfaEnabled = computed(() => store.getters["auth/isMfa"]);
const dialogMfaSettings = ref(false);
const dialogMfaDisable = ref(false);
const showChangePassword = ref(false);
const showDeleteAccountDialog = ref(false);
const getAuthMethods = computed(() => store.getters["auth/getAuthMethods"]);
const isLocalAuth = computed(() => getAuthMethods.value.includes("local"));
const { isCloud, isCommunity } = envVariables;
const { lgAndUp } = useDisplay();

const {
  value: name,
  errorMessage: nameError,
  setErrors: setNameError,
} = useField<string>("name", yup.string().required()
  .min(1, "Your name should be 1-64 characters long")
  .max(64, "Your name should be 1-64 characters long"), {
  initialValue: "",
});

const {
  value: username,
  errorMessage: usernameError,
  setErrors: setUsernameError,
} = useField<string>("name", yup.string().required()
  .min(1, "Your name should be 1-32 characters long")
  .max(32, "Your name should be 1-32 characters long")
  .test(
    "username-error",
    "The username only accepts the lowercase letters and this special characters _, ., - and @.",
    (value) => {
      const regex = /^[a-z0-9_.@-\s]*$/;
      return regex.test(value || "");
    },
  )
  .test(
    "white-spaces",
    "The username cannot contain white spaces.",
    (value) => {
      const regex = /\s/;
      return !regex.test(value || "");
    },
  ), {
  initialValue: "",
});

const {
  value: email,
  errorMessage: emailError,
  setErrors: setEmailError,
} = useField<string>("email", yup.string().email().required(), {
  initialValue: "",
});

const {
  value: recoveryEmail,
  errorMessage: recoveryEmailError,
  setErrors: setRecoveryEmailError,
} = useField<string>(
  "recoveryEmail",
  yup
    .string()
    .email()
    .test(
      "not-same-as-email",
      "Recovery email must not be the same as email",
      (value) => value !== email.value,
    ),
  {
    initialValue: "",
  },
);

const setUserData = () => {
  name.value = store.getters["auth/currentName"];
  username.value = store.getters["auth/currentUser"];
  email.value = store.getters["auth/email"];
  recoveryEmail.value = store.getters["auth/recoveryEmail"];
};

const toggleMfa = () => {
  if (mfaEnabled.value) {
    dialogMfaDisable.value = true;
  } else {
    dialogMfaSettings.value = true;
  }
};

const hasUserDataError = computed(() => nameError.value || usernameError.value || emailError.value);

const enableEdit = (form: string) => {
  if (form === "data") {
    editDataStatus.value = !editDataStatus.value;
  } else if (form === "password") {
    editPasswordStatus.value = !editPasswordStatus.value;
  }
};

const handleUpdateUserDataError = (
  error: unknown,
  setFieldError: Record<string, (msg: string) => void>,
) => {
  if (axios.isAxiosError(error)) {
    const axiosError = error as AxiosError<{ data: ErrorResponseData }>;
    const responseStatus = axiosError.response?.status;

    if (responseStatus === 409 || responseStatus === 400) {
      const errorMessages = axiosError.response?.data;
      if (Array.isArray(errorMessages)) {
        errorMessages.forEach((field) => {
          const setError = setFieldError[field];
          if (setError) {
            setError(
              responseStatus === 409
                ? `This ${field} already exists`
                : `This ${field} is invalid!`,
            );
          } else {
            console.warn(`No error handler defined for field: ${field}`);
          }
        });
      }
    } else {
      store.dispatch("snackbar/showSnackbarErrorDefault");
      handleError(error);
    }
  } else {
    store.dispatch("snackbar/showSnackbarErrorDefault");
    handleError(error);
  }
};

const updateUserData = async () => {
  if (!hasUserDataError.value) {
    const data = {
      name: name.value,
      username: username.value,
      email: email.value,
      recovery_email: recoveryEmail.value,
    };

    try {
      await store.dispatch("users/patchData", data);
      store.dispatch("auth/changeUserData", data);
      store.dispatch(
        "snackbar/showSnackbarSuccessAction",
        INotificationsSuccess.profileData,
      );
      enableEdit("data");
    } catch (error: unknown) {
      handleUpdateUserDataError(
        error,
        {
          username: setUsernameError,
          name: setNameError,
          email: setEmailError,
          recovery_email: setRecoveryEmailError,
        },
      );
    }
  }
};

const cancel = (type: string) => {
  if (type === "data") {
    setUserData();
    editDataStatus.value = !editDataStatus.value;
  }
};

onMounted(() => {
  store.dispatch("auth/getUserInfo");
  setUserData();
});
</script>

<style lang="scss" scoped>
.v-container {
  max-width: 960px;
  margin-left: 0;
  padding: 0;
}

:deep(.v-field--variant-plain) {
  --v-field-padding-start: 16px;
  --v-field-padding-end: 16px;
  --v-field-padding-bottom: 8px;
}
</style>
