<template>
  <v-app>
    <v-main>
      <v-container class="full-height d-flex justify-center align-center" fluid>
        <v-row align="center" justify="center">
          <v-col cols="12" sm="8" md="4">
            <v-alert
              v-if="!isEnterprise"
              type="success"
              variant="tonal"
              class="paywall-banner mb-4"
            >
              <template v-slot:prepend>
                <div class="circle-one shadow d-flex justify-center align-center">
                  <div class="circle-two shadow d-flex justify-center align-center">
                    <v-icon color="success" class="green-inner-shadow" size="50">
                      mdi-crown-circle
                    </v-icon>
                  </div>
                </div>
              </template>
              <template v-slot:text>
                <strong>Unlock Advanced Features with ShellHub Enterprise!</strong>
                <p class="mb-0 text-body-2">
                  Gain access to real-time session recording, role-based access control (RBAC), audit logs,
                  and priority support. Take your infrastructure to the next level!
                </p>
              </template>
            </v-alert>

            <v-btn
              v-if="!isEnterprise"
              color="success"
              block
              class="mb-4"
              variant="tonal"
              href="https://www.shellhub.io/pricing"
              rel="noopener noreferrer"
              target="_blank"
            >
              Upgrade to Enterprise Now
            </v-btn>
            <v-card theme="dark" class="pa-6" rounded="lg">
              <v-card-title class="d-flex justify-center align-center mt-4">
                <v-img
                  :src="Logo"
                  max-width="220"
                  alt="ShellHub Logo"
                />
                <span class="mt-6 text-overline">Admin</span>
              </v-card-title>
              <v-card-item>
                <SnackbarComponent />
                <form @submit.prevent="login">
                  <v-container>
                    <v-text-field
                      color="primary"
                      prepend-icon="mdi-account"
                      v-model="username"
                      :error-messages="usernameError"
                      required
                      label="Username"
                      variant="underlined"
                      data-test="username-text"
                      :disabled="!isEnterprise"
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
                      :disabled="!isEnterprise"

                    />
                    <v-card-actions class="justify-center">
                      <v-btn
                        type="submit"
                        data-test="login-btn"
                        color="primary"
                        variant="tonal"
                        block
                        @click="login"
                        :disabled="!isEnterprise"
                      >
                        LOGIN
                      </v-btn>
                    </v-card-actions>
                  </v-container>
                </form>
              </v-card-item>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-main>
  </v-app>

  <PaywallDialog
    v-model="isEnterprise"
    filter="enterprise"
    title="Unlock Full Managing with ShellHub Enterprise!"
    subtitle="Upgrade to ShellHub Enterprise (Managed or On-Premises)
    and gain full control over your infrastructure with advanced security,
    management and priority support."
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import { useRoute, useRouter } from "vue-router";
import PaywallDialog from "@global/components/User/PaywallDialog.vue";
import { useStore } from "../store";
import Logo from "../assets/logo-inverted.png";
import { createNewClient } from "../api/http";
import { envVariables } from "@/envVariables";

const showPassword = ref(false);
const store = useStore();
const route = useRoute();
const router = useRouter();
const isEnterprise = ref(!envVariables.isEnterprise);

const { value: username, errorMessage: usernameError } = useField<string | undefined>(
  "name",
  yup.string().required(),
  { initialValue: "" },
);

const { value: password, errorMessage: passwordError } = useField<string | undefined>(
  "password",
  yup.string().required(),
  { initialValue: "" },
);

// const required = (value: string) => !!value || "Required.";

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
      createNewClient();
      if (route.query.redirect) {
        router.push(`${route.query.redirect}`);
      } else {
        router.push("/");
      }
    } catch (error) {
      store.dispatch("snackbar/showSnackbarErrorDefault");
    }
  } else {
    store.dispatch("snackbar/showSnackbarErrorDefault");
  }
};
</script>

<style>
.full-height {
  height: 100vh;
}

.v-field__append-inner {
  cursor: pointer;
}
</style>
