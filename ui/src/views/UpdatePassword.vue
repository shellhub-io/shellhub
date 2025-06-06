<template>
  <v-container>
    <v-card-title class="d-flex justify-center" data-test="title">
      Reset your password
    </v-card-title>

    <v-card-text data-test="sub-text">
      <div class="d-flex align-center justify-center text-center">
        Please insert your new password.
      </div>
    </v-card-text>

    <v-card-text>
      <v-text-field
        id="password"
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
        id="password-confirm"
        color="primary"
        prepend-icon="mdi-lock"
        :append-inner-icon="showConfirmPassword ? 'mdi-eye' : 'mdi-eye-off'"
        v-model="passwordConfirm"
        :error-messages="passwordConfirmError"
        label="Confirm Password"
        required
        variant="underlined"
        data-test="password-confirm-text"
        :type="showConfirmPassword ? 'text' : 'password'"
        @click:append-inner="showConfirmPassword = !showConfirmPassword"
      />
    </v-card-text>

    <v-card-actions class="justify-center">
      <v-btn
        type="submit"
        color="primary"
        variant="tonal"
        data-test="update-password-btn"
        @click="updatePassword"
      >
        UPDATE PASSWORD
      </v-btn>
    </v-card-actions>

    <v-card-subtitle
      class="d-flex align-center justify-center pa-4 mx-auto pt-2"
      data-test="isCloud-card"
    >
      Back to
      <router-link
        class="ml-1"
        :to="{ name: 'Login' }"
      >
        Login
      </router-link>
    </v-card-subtitle>
  </v-container>
</template>

<script setup lang="ts">
import { useField } from "vee-validate";
import { onMounted, ref, watch } from "vue";
import { LocationQueryValue, useRoute, useRouter } from "vue-router";
import * as yup from "yup";
import { useStore } from "../store";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";

type TUpdatePassword = {
  id: LocationQueryValue | LocationQueryValue[];
  token: LocationQueryValue | LocationQueryValue[];
  password: string;
};

const store = useStore();
const route = useRoute();
const router = useRouter();
const snackbar = useSnackbar();
const data = ref({} as TUpdatePassword);
const showPassword = ref(false);
const showConfirmPassword = ref(false);

const {
  value: password,
  errorMessage: passwordError,
  setErrors: setPasswordError,
} = useField<string>(
  "password",
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
  value: passwordConfirm,
  errorMessage: passwordConfirmError,
  resetField: resetPasswordConfirm,
  setErrors: setPasswordConfirmError,
} = useField<string>(
  "passwordConfirm",
  yup
    .string()
    .required()
    .test(
      "passwords-match",
      "Passwords do not match",
      (value) => password.value === value,
    ),
  {
    initialValue: "",
  },
);

onMounted(() => {
  data.value = {
    id: route.query.id,
    token: route.query.token,
    password: "",
  };
});

watch(password, () => {
  if (password.value === passwordConfirm.value) {
    resetPasswordConfirm();
  }

  if (password.value !== passwordConfirm.value && passwordConfirm.value) {
    setPasswordConfirmError("Passwords do not match");
  }
});

const hasErros = () => {
  if (password.value === "") {
    setPasswordError("this is a required field");
    return true;
  }

  if (passwordConfirm.value === "") {
    setPasswordConfirmError("this is a required field");
    return true;
  }

  if (passwordError.value) {
    return true;
  }

  if (passwordConfirmError.value) {
    return true;
  }

  return false;
};

const updatePassword = async () => {
  if (hasErros()) return;
  try {
    data.value = {
      ...data.value,
      password: password.value,
    };
    await store.dispatch("users/updatePassword", data.value);
    await router.push({ name: "Login" });
    snackbar.showSuccess("Password updated successfully.");
  } catch (error: unknown) {
    snackbar.showError("Failed to update password.");
    handleError(error);
  }
};

defineExpose({
  updatePassword,
  handleError,
});
</script>
