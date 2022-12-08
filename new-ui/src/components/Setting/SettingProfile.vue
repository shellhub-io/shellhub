<template>
  <v-container>
    <v-row align="center" justify="center" class="mt-4">
      <v-col sm="8">
        <v-row>
          <v-col>
            <h3>Account</h3>
          </v-col>

          <v-spacer />

          <v-col md="auto" class="ml-auto">
            <v-btn
              v-if="!editDataStatus"
              color="primary"
              @click="editDataStatus = !editDataStatus"
            >
              Change Data
            </v-btn>

            <div v-if="editDataStatus" class="d-flex align-center">
              <v-btn class="mr-2" color="primary" @click="cancel('data')">
                Cancel
              </v-btn>

              <v-btn color="primary" @click="updateUserData"> Save </v-btn>
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
          <v-col>
            <h3>Password</h3>
          </v-col>

          <v-spacer />

          <v-col md="auto" class="ml-auto">
            <v-btn
              v-if="!editPasswordStatus"
              color="primary"
              @click="editPasswordStatus = !editPasswordStatus"
            >
              Change Password
            </v-btn>

            <div v-if="editPasswordStatus" class="d-flex align-center">
              <v-btn class="mr-2" color="primary" @click="cancel('password')">
                Cancel
              </v-btn>

              <v-btn 
                color="primary" 
                @click="updatePassword"
                :disabled="currentPasswordError !== '' || newPasswordConfirmError !== ''"
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

<script lang="ts">
import { defineComponent, ref, computed, onMounted } from "vue";
import { useStore } from "../../store";
import { useField } from "vee-validate";
import * as yup from "yup";
import { INotificationsSuccess } from "../../interfaces/INotifications";

export default defineComponent({
  setup() {
    const store = useStore();
    const editDataStatus = ref(false);
    const editPasswordStatus = ref(false);
    const show = ref(false);
    const showCurrentPassword = ref(false);
    const showNewPassword = ref(false);
    const showConfirmPassword = ref(false);

    const {
      value: name,
      errorMessage: nameError,
      setErrors: setNameError,
    } = useField<string>("name", yup.string().required(), {
      initialValue: "",
    });

    const {
      value: username,
      errorMessage: usernameError,
      setErrors: setUsernameError,
    } = useField<string>("username", yup.string().required(), {
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
      value: currentPassword,
      errorMessage: currentPasswordError,
      setErrors: setCurrentPasswordError,
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
      yup.string().required().min(5).max(30),
      {
        initialValue: "",
      }
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
          (value) => newPassword.value === value
        ),
      {
        initialValue: "",
      }
    );

    onMounted(() => {
      setUserData();
    });

    const setUserData = () => {
      name.value = store.getters["auth/currentName"];
      username.value = store.getters["auth/currentUser"];
      email.value = store.getters["auth/email"];
    };

    const hasUserDataError = computed(() => {
      return nameError.value || usernameError.value || emailError.value;
    });

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
            INotificationsSuccess.profileData
          );
          enableEdit("data");
        } catch (error: any) {
          if (error.code === 409) {
            error.body.forEach((field: string) => {
              if (field === "username")
                setUsernameError("This username already exists");
              else if (field === "name")
                setNameError("This name already exists");
              else if (field === "email")
                setEmailError("This email already exists");
            });
          } else if (error.code === 400) {
            error.body.forEach((field: string) => {
              if (field === "username")
                setUsernameError("This username is invalid !");
              else if (field === "name") setNameError("This name is invalid !");
              else if (field === "email")
                setEmailError("This email is invalid !");
            });
          } else {
            store.dispatch("snackbar/showSnackbarErrorDefault");
          }
        }
      }
    };

    const hasUpdatePasswordError = computed(() => {
      return (
        currentPasswordError.value ||
        newPasswordError.value ||
        newPasswordConfirmError.value
      );
    });

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
            INotificationsSuccess.profilePassword
          );
          enableEdit("password");
          resetPasswordFields();
        } catch (error: any) {
          if (error.response.status === 403) {
            // failed password
            setNewPasswordError("Your password doesn't match");
            setNewPasswordConfirmError("Your password doesn't match");
          } else {
            store.dispatch("snackbar/showSnackbarErrorDefault");
          }
        }
      }
    };

    const cancel = (type: string) => {
      if (type === "data") {
        setUserData();
        // this.$refs.obs.reset();
        editDataStatus.value = !editDataStatus.value;
      } else if (type === "password") {
        currentPassword.value = "";
        newPassword.value = "";
        newPasswordConfirm.value = "";
        // $refs.pass.reset();
        editPasswordStatus.value = !editPasswordStatus.value;
      }
    };

    return {
      editDataStatus,
      name,
      nameError,
      username,
      usernameError,
      email,
      emailError,
      editPasswordStatus,
      currentPassword,
      currentPasswordError,
      newPassword,
      newPasswordError,
      newPasswordConfirm,
      newPasswordConfirmError,
      show,
      showCurrentPassword,
      showNewPassword,
      showConfirmPassword,
      updateUserData,
      updatePassword,
      cancel,
    };
  },
});
</script>
