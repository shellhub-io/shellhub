<template>
  <v-container
    fluid
    data-test="account-profile-container"
  >
    <UserDeleteWarning
      v-if="showUserDeleteWarning"
      v-model="showDeleteAccountDialog"
      data-test="delete-user-community-dialog"
    />
    <UserDelete
      v-else
      v-model="showDeleteAccountDialog"
      data-test="delete-user-dialog"
    />
    <PageHeader
      icon="mdi-account-circle"
      title="Account Profile"
      overline="Settings"
      description="Manage your personal account information, authentication settings, and security preferences."
      icon-color="primary"
      data-test="account-profile-card"
    >
      <template #actions>
        <v-btn
          v-if="!editDataStatus"
          color="primary"
          variant="elevated"
          data-test="edit-profile-button"
          @click="editDataStatus = !editDataStatus"
        >
          Edit Profile
        </v-btn>
        <template v-else>
          <v-btn
            color="primary"
            variant="text"
            class="mr-2"
            data-test="cancel-edit-button"
            @click="cancel('data')"
          >
            Cancel
          </v-btn>
          <v-btn
            color="primary"
            variant="flat"
            data-test="save-changes-button"
            @click="updateUserData"
          >
            Save Changes
          </v-btn>
        </template>
      </template>
    </PageHeader>

    <SettingsSection data-test="profile-details-list">
      <SettingsRow
        icon="mdi-badge-account"
        title="Name"
        title-test-id="name-field"
      >
        <v-text-field
          v-model="name"
          :error-messages="nameError"
          :disabled="!editDataStatus"
          :readonly="!editDataStatus"
          required
          :hide-details="!nameError"
          density="compact"
          :variant="editDataStatus ? 'outlined' : 'plain'"
          :reverse="smAndUp"
          data-test="name-input"
        />
      </SettingsRow>
      <v-divider />
      <SettingsRow
        v-if="isLocalAuth || isCloud"
        icon="mdi-account"
        title="Username"
        title-test-id="username-field"
      >
        <v-text-field
          v-model="username"
          :error-messages="usernameError"
          :disabled="!editDataStatus"
          :readonly="!editDataStatus"
          density="compact"
          :variant="editDataStatus ? 'outlined' : 'plain'"
          required
          :hide-details="!usernameError"
          :reverse="smAndUp"
          data-test="username-input"
        />
      </SettingsRow>
      <v-divider v-if="isLocalAuth || isCloud" />
      <SettingsRow
        icon="mdi-email"
        title="Email"
        title-test-id="email-field"
      >
        <v-text-field
          v-model="email"
          :error-messages="emailError"
          :disabled="!editDataStatus"
          :readonly="!editDataStatus"
          density="compact"
          :variant="editDataStatus ? 'outlined' : 'plain'"
          required
          :hide-details="!emailError"
          :reverse="smAndUp"
          data-test="email-input"
        />
      </SettingsRow>
      <v-divider />
      <SettingsRow
        v-if="isLocalAuth || isCloud"
        icon="mdi-email-lock"
        title="Recovery Email"
        title-test-id="recovery-email-field"
      >
        <v-text-field
          v-model="recoveryEmail"
          :error-messages="recoveryEmailError"
          :disabled="!editDataStatus"
          :readonly="!editDataStatus"
          density="compact"
          :variant="editDataStatus ? 'outlined' : 'plain'"
          required
          :hide-details="!recoveryEmailError"
          :reverse="smAndUp"
          data-test="recovery-email-input"
        />
      </SettingsRow>
      <v-divider v-if="isLocalAuth || isCloud" />
      <SettingsRow
        v-if="isLocalAuth || isCloud"
        icon="mdi-key"
        title="Password"
      >
        <v-btn
          variant="text"
          color="primary"
          @click="showChangePassword = true"
        >
          Change Password
        </v-btn>
        <ChangePassword v-model="showChangePassword" />
      </SettingsRow>
      <v-divider v-if="isLocalAuth || isCloud" />
      <SettingsRow
        v-if="isLocalAuth || isCloud"
        icon="mdi-fingerprint"
        title="Multi-factor Authentication"
        subtitle="Enable multi-factor authentication (MFA) to add an extra layer of security to your account.
         You'll need to enter a one-time verification code from your preferred TOTP provider to log in."
        subtitle-test-id="mfa-text"
        data-test="mfa-card"
      >
        <v-tooltip
          location="top"
          text="Only available for Cloud or Enterprise accounts!"
          :disabled="!isCommunity"
        >
          <template #activator="{ props }">
            <div
              v-bind="props"
              class="d-flex align-center"
            >
              <v-switch
                v-model="isMfaEnabled"
                class="mr-sm-4"
                hide-details
                inset
                color="primary"
                :disabled="isCommunity"
                data-test="switch-mfa"
                @click.prevent="toggleMfa()"
              />
              <MfaSettings v-model="showMfaSettingsDialog" />
              <MfaDisable v-model="showMfaDisableDialog" />
            </div>
          </template>
        </v-tooltip>
      </SettingsRow>
      <v-divider v-if="isLocalAuth || isCloud" />
      <SettingsRow
        icon="mdi-delete"
        title="Delete Account"
        title-test-id="delete-account"
      >
        <v-btn
          variant="text"
          color="error"
          data-test="delete-account-btn"
          @click="showDeleteAccountDialog = true"
        >
          Delete
        </v-btn>
      </SettingsRow>
    </SettingsSection>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useDisplay } from "vuetify";
import { useField } from "vee-validate";
import axios, { AxiosError } from "axios";
import * as yup from "yup";
import handleError from "@/utils/handleError";
import MfaSettings from "../AuthMFA/MfaSettings.vue";
import MfaDisable from "../AuthMFA/MfaDisable.vue";
import UserDelete from "../User/UserDelete.vue";
import UserDeleteWarning from "../User/UserDeleteWarning.vue";
import PageHeader from "../PageHeader.vue";
import { envVariables } from "@/envVariables";
import ChangePassword from "../User/ChangePassword.vue";
import SettingsRow from "./SettingsRow.vue";
import SettingsSection from "./SettingsSection.vue";
import useSnackbar from "@/helpers/snackbar";
import useAuthStore from "@/store/modules/auth";
import useUsersStore from "@/store/modules/users";

type ErrorResponseData = { field: string; message: string }[];

const authStore = useAuthStore();
const usersStore = useUsersStore();
const snackbar = useSnackbar();
const editDataStatus = ref(false);
const editPasswordStatus = ref(false);
const isMfaEnabled = computed(() => authStore.isMfaEnabled);
const showMfaSettingsDialog = ref(false);
const showMfaDisableDialog = ref(false);
const showChangePassword = ref(false);
const showDeleteAccountDialog = ref(false);
const authMethods = computed(() => authStore.authMethods);
const isLocalAuth = computed(() => authMethods.value.includes("local"));
const { isEnterprise, isCommunity, isCloud } = envVariables;
const showUserDeleteWarning = computed(() => isCommunity || isEnterprise);
const { smAndUp } = useDisplay();

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

const toggleMfa = () => {
  if (isMfaEnabled.value) showMfaDisableDialog.value = true;
  else showMfaSettingsDialog.value = true;
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
          const setError = setFieldError[field as string] as ((msg: string) => void);
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
      snackbar.showError("An error occurred while updating user data.");
      handleError(error);
    }
  } else {
    snackbar.showError("An error occurred while updating user data.");
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
      await usersStore.patchData(data);
      authStore.updateUserData(data);
      snackbar.showSuccess("Profile data updated successfully.");
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

const setUserData = () => {
  name.value = authStore.name;
  username.value = authStore.username;
  email.value = authStore.email;
  recoveryEmail.value = authStore.recoveryEmail;
};

const cancel = (type: string) => {
  if (type === "data") {
    setUserData();
    editDataStatus.value = !editDataStatus.value;
  }
};

onMounted(async () => {
  await authStore.getUserInfo();
  setUserData();
});
</script>

<style lang="scss" scoped>
:deep(.v-field--variant-plain) {
  --v-field-padding-start: 16px;
  --v-field-padding-end: 16px;
  --v-field-padding-bottom: 8px;
}

@media (max-width: 600px) {
  :deep(.v-text-field .v-field__input) {
    text-align: center;
  }
}

</style>
