<template>
  <v-app>
    <v-main class="d-flex align-center justify-center">
      <v-container fluid>
        <v-row align="center" justify="center">
          <v-col cols="12" sm="8" md="4">
            <v-card
              v-if="showMessage && !isCloud"
              class="bg-v-theme-surface"
              data-test="unknownReason-card"
            >
              <v-card-text>
                <v-card-title class="justify-center">
                  Activate Account
                </v-card-title>

                <div class="d-flex align-center justify-center mb-6">
                  The account is not active for an unknown reason.
                </div>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>

        <v-row align="center" justify="center">
          <v-col cols="12" sm="8" md="4">
            <v-alert
              v-if="loginToken"
              data-test="loadingToken-alert"
              class="pa-6 bg-v-theme-surface"
            >
              <div class="text-center">
                <p>Logging the token in...</p>

                <v-progress-linear indeterminate color="cyan" class="mt-2" />
              </div>
            </v-alert>
          </v-col>
        </v-row>

        <v-row align="center" justify="center">
          <v-col cols="12" sm="8" md="4">
            <v-card theme="dark" class="pa-6 bg-v-theme-surface" rounded="lg">
              <v-card-title class="d-flex justify-center align-center mt-4">
                <v-img
                  :src="Logo"
                  max-width="220"
                  alt="logo do ShellHub, uma nuvem de com a escrita ShellHub Admin ao lado"
                />
              </v-card-title>
              <v-container>
                <SnackbarComponent />
                <form @submit.prevent="login">
                  <v-container>
                    <v-text-field
                      color="primary"
                      prepend-icon="mdi-account"
                      v-model="username"
                      :error-messages="usernameError"
                      required
                      label="Username or email address"
                      variant="underlined"
                      data-test="username-text"
                    />

                    <v-text-field
                      color="primary"
                      prepend-icon="mdi-lock"
                      :append-inner-icon="
                        showPassword ? 'mdi-eye' : 'mdi-eye-off'
                      "
                      v-model="password"
                      :error-messages="passwordError"
                      label="Password"
                      required
                      variant="underlined"
                      data-test="password-text"
                      :type="showPassword ? 'text' : 'password'"
                      @click:append-inner="showPassword = !showPassword"
                    />
                    <v-card-actions class="justify-center">
                      <v-btn
                        type="submit"
                        data-test="login-btn"
                        color="primary"
                        variant="tonal"
                        block
                        @click="login"
                      >
                        LOGIN
                      </v-btn>
                    </v-card-actions>

                    <v-card-subtitle
                      v-if="isCloud"
                      class="d-flex align-center justify-center pa-4 mx-auto pt-4 pb-0"
                      data-test="forgotPassword-card"
                    >
                      Forgot your
                      <router-link
                        class="ml-1"
                        :to="{ name: 'ForgotPassword' }"
                      >
                        Password?
                      </router-link>
                    </v-card-subtitle>

                    <v-card-subtitle
                      v-if="isCloud"
                      class="d-flex align-center justify-center pa-4 mx-auto"
                      data-test="isCloud-card"
                    >
                      Don't have an account?

                      <router-link class="ml-1" :to="{ name: 'SignUp' }">
                        Sign up here
                      </router-link>
                    </v-card-subtitle>
                  </v-container>
                </form>
              </v-container>
            </v-card>
          </v-col>
        </v-row>

        <v-row align="center" justify="center">
          <v-col cols="12" sm="8" md="4">
            <AccountCreated
              v-if="isCloud"
              :show="showMessage"
              :username="username"
              data-test="accountCreated-component"
            />
          </v-col>
        </v-row>
      </v-container>
    </v-main>
  </v-app>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, ref } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import { useRoute, useRouter } from "vue-router";
import { useStore } from "../store";
import Logo from "../assets/logo-inverted.png";
import { envVariables } from "../envVariables";
import { createNewClient } from "../api/http";
import { INotificationsError } from "../interfaces/INotifications";
import AccountCreated from "../components/Account/AccountCreated.vue";

export default defineComponent({
  name: "Login",
  setup() {
    const showPassword = ref(false);
    const loginToken = ref(false);
    const showMessage = ref(false);
    const store = useStore();
    const route = useRoute();
    const router = useRouter();
    const isCloud = computed(() => envVariables.isCloud);
    const hasNamespace = computed(
      () => store.getters["namespaces/getNumberNamespaces"] !== 0,
    );
    const isTheSameNamespace = computed(() => store.getters["namespaces/get"].tenant_id === localStorage.getItem("tenant"));

    onMounted(async () => {
      if (route.query.token) {
        store.dispatch("layout/setLayout", "simpleLayout");
        loginToken.value = true;
        await store.dispatch("stats/clear");
        await store.dispatch("namespaces/clearNamespaceList");
        await store.dispatch("auth/logout");
        createNewClient();

        store.dispatch("auth/loginToken", route.query.token);
        await store.dispatch("auth/loginToken", route.query.token).then(async () => {
          createNewClient();
          await router.push("/");
          store.dispatch("layout/setLayout", "appLayout");
        });
      }
    });

    const { value: username, errorMessage: usernameError } = useField<string>(
      "name",
      yup.string().required(),
      { initialValue: "" },
    );
    const { value: password, errorMessage: passwordError } = useField<string>(
      "password",
      yup.string().required(),
      { initialValue: "" },
    );
    const required = (value: string) => !!value || "Required.";

    const hasErrors = () => {
      if (usernameError.value || passwordError.value) {
        return true;
      }
      return false;
    };

    const login = async () => {
      if (!hasErrors() && username.value && password.value) {
        try {
          await store.dispatch("auth/login", {
            username: username.value,
            password: password.value,
          });
          await createNewClient();
          await store.dispatch("layout/setLayout", "appLayout");

          if (hasNamespace.value && !isTheSameNamespace.value) {
            await store.dispatch("namespaces/get", localStorage.getItem("tenant"));
          }

          if (route.query.redirect) {
            router.push(`${route.query.redirect}`);
          } else {
            router.push("/");
          }
        } catch (error: any) {
          switch (true) {
            case error.response.status === 401: {
              store.dispatch(
                "snackbar/showSnackbarErrorIncorrect",
                INotificationsError.loginFailed,
              );
              break;
            }
            case error.response.status === 403: {
              showMessage.value = !showMessage.value;
              break;
            }
            default: {
              store.dispatch("snackbar/showSnackbarErrorDefault");
              throw new Error(error);
            }
          }
        }
      } else {
        store.dispatch("snackbar/showSnackbarErrorDefault");
      }
    };

    return {
      Logo,
      username,
      usernameError,
      password,
      passwordError,
      showPassword,
      showMessage,
      loginToken,
      required,
      isCloud,
      store,
      login,
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
