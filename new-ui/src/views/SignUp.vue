<template>
  <v-app>
    <v-main class="d-flex justify-center align-center">
      <v-container class="w-auto" fluid>
        <div class="text-center">
          <v-overlay :value="overlay">
            <v-progress-circular
              indeterminate
              size="64"
              alt="Sign Up loading"
            />
          </v-overlay>
        </div>
        <v-card
          v-if="!showMessage"
          theme="dark"
          class="pa-6 bg-v-theme-surface"
          rounded="lg"
        >
          <v-card-title class="d-flex justify-center align-center">
            <v-img
              :src="Logo"
              max-width="220"
              alt="ShellHub logo, a cloud with a shell in your base write ShellHub in the right side"
            />
          </v-card-title>
          <v-container class="pb-0 mb-0">
            <form @submit.prevent="createAccount">
              <v-card-title class="text-center">Create Account</v-card-title>
              <v-container>
                <v-text-field
                  color="primary"
                  prepend-icon="mdi-account"
                  v-model="name"
                  :error-messages="nameError"
                  required
                  label="Name"
                  variant="underlined"
                  data-test="name-text"
                />

                <v-text-field
                  color="primary"
                  prepend-icon="mdi-account"
                  v-model="username"
                  :error-messages="usernameError"
                  required
                  label="Username"
                  variant="underlined"
                  data-test="username-text"
                />

                <v-text-field
                  color="primary"
                  prepend-icon="mdi-email"
                  v-model="email"
                  :error-messages="emailError"
                  required
                  label="Email"
                  variant="underlined"
                  data-test="email-text"
                />

                <v-text-field
                  color="primary"
                  prepend-icon="mdi-lock"
                  :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
                  v-model="password"
                  :error-messages="passwordError"
                  label="Password"
                  required
                  variant="underlined"
                  data-test="password-text"
                  :type="showPassword ? 'text' : 'password'"
                  @click:append-inner="showPassword = !showPassword"
                />

                <v-text-field
                  color="primary"
                  prepend-icon="mdi-lock"
                  :append-inner-icon="
                    showConfirmPassword ? 'mdi-eye' : 'mdi-eye-off'
                  "
                  v-model="passwordConfirm"
                  :error-messages="passwordConfirmError"
                  label="Confirm Password"
                  required
                  variant="underlined"
                  data-test="password-confirm-text"
                  :type="showConfirmPassword ? 'text' : 'password'"
                  @click:append-inner="
                    showConfirmPassword = !showConfirmPassword
                  "
                />
              </v-container>

              <div v-if="isCloud">
                <v-checkbox
                  v-model="acceptPrivacyPolicy"
                  color="primary"
                  hide-details
                  data-test="accept-privacy-policy-checkbox"
                >
                  <template #label>
                    <span class="caption">
                      I agree to the
                      <a
                        href="https://website-git-fork-antonyrafael-feat-privacy-policy-page-shellhub.vercel.app/privacy-policy"
                        target="_blank"
                        rel="noopener noreferrer"
                        >Privacy Policy</a
                      >
                    </span>
                  </template>
                </v-checkbox>
                <v-checkbox
                v-model="acceptMarketing"
                color="primary"
                hide-details
                data-test="accept-news-checkbox"
                >
                  <template #label>
                    <p>
                      I accept to receive news and updates from ShellHub via
                      email.
                    </p>
                  </template>
                </v-checkbox>
              </div>
              <v-card-subtitle
                v-if="privacyPolicyError"
                class="pa-0 pl-2 font-weight-medium text-error"
                data-test="privacy-policy-error"
              >
                You need to accept the Privacy Policy to create an account.
              </v-card-subtitle>

              <v-card-actions class="justify-center">
                <v-btn
                  type="submit"
                  data-test="login-btn"
                  color="primary"
                  variant="tonal"
                  block
                >
                  CREATE
                </v-btn>

              </v-card-actions>

              <v-card-subtitle
                class="d-flex align-center justify-center pa-4 mx-auto"
                data-test="isCloud-card"
              >
                Do you have account ?
                <router-link class="ml-1" :to="{ name: 'login' }">
                  Login
                </router-link>
              </v-card-subtitle>
            </form>
          </v-container>
        </v-card>
        <AccountCreated
          :show="showMessage"
          :username="username"
          data-test="accountCreated-component"
        />
      </v-container>
    </v-main>
  </v-app>
</template>

<script lang="ts">
import { computed, defineComponent, ref, watch } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import Logo from "../assets/logo-inverted.png";
import { useStore } from "../store";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../interfaces/INotifications";
import AccountCreated from "../components/Account/AccountCreated.vue";
import { envVariables } from "@/envVariables";

export default defineComponent({
  setup() {
    const store = useStore();
    const showPassword = ref(false);
    const showConfirmPassword = ref(false);
    const showMessage = ref(false);
    const acceptMarketing = ref(false);
    const acceptPrivacyPolicy = ref(false);
    const privacyPolicyError = ref(false);
    const delay = ref(500);
    const overlay = ref(false);
    const isCloud = computed(() => envVariables.isCloud);

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
    } = useField<string>(
      "username",
      yup
        .string()
        .required()
        .min(3)
        .max(30)
        .test(
          "username-error",
          "The username only accepts the special characters _, ., - and @.",
          (value) => {
            const regex = /^[a-zA-Z0-9_.@-\s]*$/;
            // @ts-ignore
            return regex.test(value);
          }
        )
        .test(
          "white-spaces",
          "The username cannot contain white spaces.",
          (value) => {
            const regex = /\s/;
            // @ts-ignore
            return !regex.test(value);
          }
        ),
      {
        initialValue: "",
      }
    );

    const {
      value: email,
      errorMessage: emailError,
      setErrors: setEmailError,
    } = useField<string>("email", yup.string().email().required(), {
      initialValue: "",
    });

    const {
      value: password,
      errorMessage: passwordError,
      setErrors: setPasswordError,
    } = useField<string>(
      "password",
      yup
        .string()
        .required()
        .min(5, "Your password should be 5-30 characters long")
        .max(30, "Your password should be 5-30 characters long"),
      {
        initialValue: "",
      }
    );

    const {
      value: passwordConfirm,
      errorMessage: passwordConfirmError,
      setErrors: setPasswordConfirmError,
      resetField: resetPasswordConfirm,
    } = useField<string>(
      "passwordConfirm",
      yup
        .string()
        .required()
        .test(
          "passwords-match",
          "Passwords do not match",
          (value) => password.value === value
        ),
      {
        initialValue: "",
      }
    );

    watch(overlay, (value) => {
      if (value) {
        setTimeout(() => {
          overlay.value = false;
          store.dispatch(
            "snackbar/showSnackbarSuccessAction",
            INotificationsSuccess.addUser
          );
        }, delay.value);
      }
    });

    watch(acceptPrivacyPolicy, (value) => {
      if (value) {
        privacyPolicyError.value = false;
      }
    });

    const hasErrors = () => {
      if (
        nameError.value ||
        usernameError.value ||
        emailError.value ||
        passwordError.value ||
        passwordConfirmError.value ||
        !name.value ||
        !username.value ||
        !email.value ||
        !password.value ||
        !passwordConfirm.value
      ) {
        return true;
      }
      return false;
    };

    const createAccount = async () => {
      if (!hasErrors()) {
        try {

          if (isCloud.value && !acceptPrivacyPolicy.value) {
            privacyPolicyError.value = true;
            return;
          }

          await store.dispatch("users/signUp", {
            name: name.value,
            email: email.value,
            username: username.value,
            password: password.value,
            confirmPassword: passwordConfirm.value,
            emailMarketing: acceptMarketing.value,
          });
          overlay.value = !overlay.value;
          showMessage.value = !showMessage.value;
          store.dispatch(
            "snackbar/showSnackbarSuccessAction",
            INotificationsSuccess.addUser
          );
        } catch (e: any) {
          store.dispatch(
            "snackbar/showSnackbarErrorAction",
            INotificationsError.addUser
          );
          if (e.response.status === 409) {
            e.response.data.forEach((field: string) => {
              if (field === "username")
                setUsernameError("This username already exists");
              else if (field === "name")
                setNameError("This name already exists");
              else if (field === "email")
                setEmailError("This email already exists");
              else if (field === "password")
                setPasswordError("This password already exists");
            });
          } else if (e.response.status === 400) {
            e.response.data.forEach((field: string) => {
              if (field === "username")
                setUsernameError("This username is invalid !");
              else if (field === "name") setNameError("This name is invalid !");
              else if (field === "email")
                setEmailError("This email is invalid !");
              else if (field === "password")
                setPasswordError("This password is invalid !");
            });
          }
        }
      }
    };
    return {
      Logo,
      showPassword,
      name,
      nameError,
      username,
      usernameError,
      email,
      emailError,
      password,
      passwordError,
      passwordConfirm,
      passwordConfirmError,
      showConfirmPassword,
      createAccount,
      store,
      showMessage,
      delay,
      overlay,
      isCloud,
      acceptMarketing,
      acceptPrivacyPolicy,
      privacyPolicyError,
    };
  },
  components: { AccountCreated },
});
</script>

<style>
.full-height {
  height: 100vh;
}

.v-field__append-inner {
  cursor: pointer;
}
</style>
