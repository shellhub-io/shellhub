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
import { useStore } from "../../store";
import { INotificationsSuccess } from "../../interfaces/INotifications";
import handleError from "@/utils/handleError";

const store = useStore();
const editDataStatus = ref(false);
const editPasswordStatus = ref(false);
const showCurrentPassword = ref(false);
const showNewPassword = ref(false);
const showConfirmPassword = ref(false);

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
} = useField<string>(
  "username",
  yup
    .string()
    .required()
    .min(3)
    .max(32)
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
    ),
  {
    initialValue: "",
  },
);

const {
  value: email,
  errorMessage: emailError,
  setErrors: setEmailError,
} = useField<string>("email", yup.string().email().required(), {
  initialValue: "",
});

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
};

onMounted(() => {
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
          });
          break;
        case axiosError.response?.status === 400:
          // @ts-expect-error axiosError.response.data is an array
          axiosError.response.data.forEach((field: string) => {
            if (field === "username") setUsernameError("This username is invalid !");
            else if (field === "name") setNameError("This name is invalid !");
            else if (field === "email") setEmailError("This email is invalid !");
          });
          break;
        default:
          store.dispatch("snackbar/showSnackbarErrorDefault");
          handleError(error);
          break;
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
</script>
