<template>
  <v-container>
    <v-row align="center" justify="center" class="mt-4">
      <v-col sm="8">
        <v-row>
          <v-col data-test="account-header">
            <h3>Account</h3>
          </v-col>

          <v-spacer />

          <v-col md="auto" class="ml-auto">
            <v-btn
              v-if="!editDataStatus"
              color="primary"
              data-test="change-data-btn"
              @click="editDataStatus = !editDataStatus"
            >
              Change Data
            </v-btn>

            <div v-if="editDataStatus" class="d-flex align-center">
              <v-btn data-test="cancel-btn" class="mr-2" color="primary" @click="cancel('data')">
                Cancel
              </v-btn>

              <v-btn data-test="update-user-btn" color="primary" @click="updateUserData"> Save </v-btn>
            </div>
          </v-col>
        </v-row>

        <div class="mt-4 pl-4 pr-4">
          <v-text-field
            v-model="name"
            label="Name"
            :error-messages="nameError"
            :disabled="!editDataStatus"
            required
            variant="underlined"
            data-test="name-text"
          />

          <v-text-field
            v-model="username"
            label="Username"
            :error-messages="usernameError"
            :disabled="!editDataStatus"
            required
            variant="underlined"
            data-test="username-text"
          />

          <v-text-field
            v-model="email"
            label="Email"
            :error-messages="emailError"
            :disabled="!editDataStatus"
            required
            variant="underlined"
            data-test="email-text"
          />

          <v-text-field
            v-if="envVariables.isCloud"
            v-model="recoveryEmail"
            label="Recovery Email"
            :error-messages="recoveryEmailError"
            :disabled="!editDataStatus"
            variant="underlined"
            data-test="recovery-email-text"
          />
        </div>

        <v-divider class="mt-6" />
        <v-divider class="mb-6" />

        <v-row>
          <v-col data-test="password-header">
            <h3>Password</h3>
          </v-col>

          <v-spacer />

          <v-col md="auto" class="ml-auto">
            <v-btn
              v-if="!editPasswordStatus"
              color="primary"
              @click="editPasswordStatus = !editPasswordStatus"
              data-test="change-password-btn"
            >
              Change Password
            </v-btn>

            <div v-if="editPasswordStatus" class="d-flex align-center">
              <v-btn data-test="cancel-password-btn" class="mr-2" color="primary" @click="cancel('password')">
                Cancel
              </v-btn>

              <v-btn
                color="primary"
                data-test="update-password-btn"
                @click="updatePassword"
                :disabled="hasUpdatePasswordError"
              > Save </v-btn>
            </div>
          </v-col>
        </v-row>

        <div class="mt-4 pl-4 pr-4">
          <v-text-field
            v-model="currentPassword"
            label="Current password"
            :append-icon="showCurrentPassword ? 'mdi-eye' : 'mdi-eye-off'"
            :type="showCurrentPassword ? 'text' : 'password'"
            class="mb-4"
            variant="underlined"
            :error-messages="currentPasswordError"
            required
            :disabled="!editPasswordStatus"
            data-test="password-text"
            @click:append="showCurrentPassword = !showCurrentPassword"
          />

          <v-text-field
            v-model="newPassword"
            label="New password"
            :append-icon="showNewPassword ? 'mdi-eye' : 'mdi-eye-off'"
            :type="showNewPassword ? 'text' : 'password'"
            class="mb-4"
            :error-messages="newPasswordError"
            required
            variant="underlined"
            :disabled="!editPasswordStatus"
            data-test="newPassword-text"
            @click:append="showNewPassword = !showNewPassword"
          />

          <v-text-field
            v-model="newPasswordConfirm"
            label="Confirm new password"
            :append-icon="showConfirmPassword ? 'mdi-eye' : 'mdi-eye-off'"
            :type="showConfirmPassword ? 'text' : 'password'"
            class="mb-4"
            variant="underlined"
            :error-messages="newPasswordConfirmError"
            required
            :disabled="!editPasswordStatus"
            data-test="confirmNewPassword-text"
            @click:append="showConfirmPassword = !showConfirmPassword"
          />
        </div>

        <v-divider class="mt-6" />
        <v-divider class="mb-6" />
        <div v-if="isEnterprise">
          <v-row>
            <v-col>
              <h3>
                Multi-factor Authentication
              </h3>
            </v-col>
          </v-row>

          <div class="mt-4 pl-4 pr-4 pb-4 mb-4">
            <p class="mb-4">Multi-factor authentication (MFA) requires users to enter a one-time verification code sent
              using your favorite TOPT Provider in order to access your ShellHub account.</p>
            <div v-if="mfaEnabled === 'true'">
              <mfa-disable @success="() => mfaEnabled = 'false'" />
            </div>
            <div v-else>
              <mfa-settings @enabled="showCongratulationsModal" />
            </div>
          </div>
          <v-row justify="center">
            <v-dialog v-model="dialog" width="auto" scrollable transition="dialog-bottom-transition" data-test="dialog">
              <v-card class="bg-v-theme-surface content" width="650" data-test="card-first-page">
                <v-container>
                  <v-row>
                    <v-col align="center" data-test="congratulation-text">
                      <h2>Congratulations! You've successfully verified your code.</h2>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col align="center">
                      <v-icon
                        end
                        icon="mdi-cloud-lock-outline"
                        color="green"
                        size="100"
                        class="green-cloud"
                        data-test="green-cloud-icon" />
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col align="start" class="ml-5 pb-0" data-test="title-bp">
                      <h4>Your account is now more secure with:</h4>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col class="ml-5 pt-0" data-test="congratulation-bullet-point">
                      <ul>
                        <li>Two-step verification adding an extra layer of protection.</li>
                        <li>Reduced risk of unauthorized access even if your password is compromised.</li>
                        <li>Enhanced security against phishing attacks and identity theft.</li>
                      </ul>
                    </v-col>
                  </v-row>
                  <v-card-actions>
                    <v-spacer />
                    <v-btn variant="text" data-test="close-btn" @keyup.enter="close()" @click="close()">
                      Close
                    </v-btn>
                  </v-card-actions>
                </v-container>
              </v-card>
            </v-dialog>
          </v-row>
        </div>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
/* eslint-disable @typescript-eslint/ban-ts-comment */
import { ref, computed, onMounted } from "vue";
import { useField } from "vee-validate";
import axios, { AxiosError } from "axios";
import * as yup from "yup";
import { useStore } from "@/store";
import { INotificationsSuccess } from "../../interfaces/INotifications";
import handleError from "@/utils/handleError";
import MfaSettings from "../AuthMFA/MfaSettings.vue";
import MfaDisable from "../AuthMFA/MfaDisable.vue";
import { envVariables } from "../../envVariables";

const store = useStore();
const editDataStatus = ref(false);
const editPasswordStatus = ref(false);
const showCurrentPassword = ref(false);
const showNewPassword = ref(false);
const showConfirmPassword = ref(false);
const dialog = ref(false);
const mfaEnabled = ref(computed(() => localStorage.getItem("mfa")).value);
const isEnterprise = computed(() => envVariables.isEnterprise);

const showCongratulationsModal = () => {
  mfaEnabled.value = "true";
  dialog.value = true;
};

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

const {
  value: currentPassword,
  errorMessage: currentPasswordError,
  resetField: resetCurrentPassword,
} = useField<string>("currentPassword", yup.string().required(), {
  initialValue: "",
});

const {
  value: newPassword,
  errorMessage: newPasswordError,
  setErrors: setNewPasswordError,
  resetField: resetNewPassword,
} = useField<string>(
  "newPassword",
  yup
    .string()
    .required()
    .min(5, "Your password should be 5-32 characters long")
    .max(32, "Your password should be 5-32 characters long"),
  {
    initialValue: "",
  },
);

const {
  value: newPasswordConfirm,
  errorMessage: newPasswordConfirmError,
  setErrors: setNewPasswordConfirmError,
  resetField: resetNewPasswordConfirm,
} = useField<string>(
  "newPasswordConfirm",
  yup
    .string()
    .required()
    .test(
      "passwords-match",
      "Passwords do not match",
      (value) => newPassword.value === value,
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

onMounted(async () => {
  await store.dispatch("auth/getUserInfo");
  setUserData();
});

const hasUserDataError = computed(() => nameError.value || usernameError.value || emailError.value);

const enableEdit = (form: string) => {
  if (form === "data") {
    editDataStatus.value = !editDataStatus.value;
  } else if (form === "password") {
    editPasswordStatus.value = !editPasswordStatus.value;
  }
};

const updateUserData = async () => {
  if (!hasUserDataError.value) {
    const data = {
      id: store.getters["auth/id"],
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
      const axiosError = error as AxiosError;
      switch (axios.isAxiosError(error)) {
        case axiosError.response?.status === 409:
          // @ts-expect-error axiosError.response.data is an array
          axiosError.response.data.forEach((field: string) => {
            if (field === "username") setUsernameError("This username already exists");
            else if (field === "name") setNameError("This name already exists");
            else if (field === "email") setEmailError("This email already exists");
            else if (field === "recovery_email") setRecoveryEmailError("This recovery email already exists");
          });
          break;
        case axiosError.response?.status === 400:
          // @ts-expect-error axiosError.response.data is an array
          axiosError.response.data.forEach((field: string) => {
            if (field === "username") setUsernameError("This username is invalid !");
            else if (field === "name") setNameError("This name is invalid !");
            else if (field === "email") setEmailError("This email is invalid !");
            else if (field === "recovery_email") setRecoveryEmailError("This recovery email is invalid !");
          });
          break;
        default:
          store.dispatch("snackbar/showSnackbarErrorDefault");
          handleError(error);
      }
    }
  }
};

const hasUpdatePasswordError = computed(() => (
  Boolean(currentPasswordError.value)
        || Boolean(newPasswordError.value)
        || Boolean(newPasswordConfirmError.value)
        || newPassword.value === ""
        || newPasswordConfirm.value === ""
        || currentPassword.value === ""
));

const resetPasswordFields = () => {
  resetCurrentPassword();
  resetNewPassword();
  resetNewPasswordConfirm();
};

const updatePassword = async () => {
  if (!hasUpdatePasswordError.value) {
    const data = {
      id: store.getters["auth/id"],
      currentPassword: currentPassword.value,
      newPassword: newPassword.value,
    };

    try {
      await store.dispatch("users/patchPassword", data);
      store.dispatch(
        "snackbar/showSnackbarSuccessAction",
        INotificationsSuccess.profilePassword,
      );
      enableEdit("password");
      resetPasswordFields();
    } catch (error: unknown) {
      if (axios.isAxiosError(error)) {
        const axiosError = error as AxiosError;
        if (axiosError.response?.status === 403) {
          // failed password
          setNewPasswordError("Your password doesn't match");
          setNewPasswordConfirmError("Your password doesn't match");
        }
      } else {
        store.dispatch("snackbar/showSnackbarErrorDefault");
        handleError(error);
      }
    }
  }
};

const cancel = (type: string) => {
  if (type === "data") {
    setUserData();
    editDataStatus.value = !editDataStatus.value;
  } else if (type === "password") {
    resetPasswordFields();
    editPasswordStatus.value = !editPasswordStatus.value;
  }
};

const close = () => {
  dialog.value = false;
};
</script>

<style scoped>
.green-cloud {
  filter: drop-shadow(0px 0px 30px rgba(43, 255, 10, 0.444))
}
</style>
