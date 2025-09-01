<template>
  <v-btn
    v-if="createUser"
    @click="showDialog = true"
    class="mr-2"
    outlined
    tabindex="0"
    aria-label="Dialog Add user"
    data-test="user-add-btn"
  >
    Add User
  </v-btn>

  <v-tooltip v-else bottom anchor="bottom">
    <template v-slot:activator="{ props }">
      <v-icon
        @click="showDialog = true"
        tag="button"
        dark
        v-bind="props"
        tabindex="0"
        aria-label="Dialog edit user"
      >mdi-pencil
      </v-icon>
    </template>
    <span>Edit</span>
  </v-tooltip>

  <BaseDialog v-model="showDialog" @close="close" transition="dialog-bottom-transition">
    <v-card>
      <v-card-title class="text-h5 pb-2">{{ titleCard }}</v-card-title>
      <v-divider />
      <form @submit="onSubmit">
        <v-card-text>
          <v-container>
            <v-row>
              <v-col cols="12">
                <v-container>
                  <v-text-field
                    v-model="name"
                    label="Name"
                    required
                    :error-messages="nameError"
                    color="primary"
                    variant="underlined"
                  />
                  <v-text-field
                    v-model="username"
                    label="Username"
                    required
                    :error-messages="usernameError"
                    color="primary"
                    variant="underlined"
                  />
                  <v-text-field
                    v-model="email"
                    label="Email"
                    required
                    name="email"
                    :error-messages="emailError"
                    color="primary"
                    variant="underlined"
                  />
                  <v-text-field
                    v-model="password"
                    label="Password"
                    :required="createUser"
                    name="password"
                    :error-messages="passwordError"
                    color="primary"
                    variant="underlined"
                    :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
                    @click:append-inner="togglePasswordVisibility"
                    :type="showPassword ? 'text' : 'password'"
                  />
                  <v-checkbox
                    v-model="changeNamespaceLimit"
                    label="Change the namespace creation limit for this user"
                    color="primary"
                    class="ml-n2"
                  />
                  <v-checkbox
                    v-if="changeNamespaceLimit"
                    v-model="disableNamespaceCreation"
                    label="Disable namespace creation"
                    color="primary"
                    class="ml-n2"
                    @change="setMaxNamespaces"
                  />
                  <v-number-input
                    v-if="changeNamespaceLimit"
                    :disabled="disableNamespaceCreation"
                    v-model="maxNamespaces"
                    label="Namespace limit"
                    :min="1"
                    class="ml-n2"
                    color="primary"
                    variant="outlined"
                  />
                  <v-tooltip location="bottom" class="text-center" :disabled="canChangeStatus">
                    <template v-slot:activator="{ props }">
                      <div v-bind="props">
                        <v-checkbox
                          v-if="!createUser"
                          label="User confirmed"
                          v-model="isConfirmed"
                          :disabled="!canChangeStatus"
                          density="compact"
                          hide-details
                          color="primary"
                        />
                      </div>
                    </template>
                    <span>{{ statusTooltipMessage }}</span>
                  </v-tooltip>
                </v-container>
              </v-col>
            </v-row>
          </v-container>
        </v-card-text>
        <v-card-actions class="pa-4">
          <v-spacer />
          <v-btn class="mr-2" @click="close" type="reset">Cancel</v-btn>
          <v-btn class="mr-2" color="primary" type="submit">{{ createUser ? "Create" : "Update" }}</v-btn>
        </v-card-actions>
      </form>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import axios, { AxiosError } from "axios";
import * as yup from "yup";
import { useField, useForm } from "vee-validate";
import useUsersStore from "@admin/store/modules/users";
import { IUser } from "@admin/interfaces/IUser";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "@/components/BaseDialog.vue";

const props = defineProps<{
  createUser?: boolean;
  user?: IUser;
  titleCard: string;
}>();

const showDialog = ref(false);
const showPassword = ref(false);
const changeNamespaceLimit = ref(props.user?.max_namespaces !== -1);
const disableNamespaceCreation = ref(false);
const maxNamespaces = ref(props.user?.max_namespaces || 0);
const canChangeStatus = props.user?.status === "not-confirmed"; // Only allow changing status if the user is not confirmed
const snackbar = useSnackbar();
const userStore = useUsersStore();
const statusTooltipMessage = props.user?.status === "invited"
  ? "You cannot change the status of an invited user."
  : "You cannot remove confirmation from a user.";

const { value: name,
  errorMessage: nameError,
  resetField: resetName,
} = useField<string | undefined>("name", yup.string().required(), {
  initialValue: props.user?.name,
});

const { value: email,
  errorMessage: emailError,
  resetField: resetEmail,
} = useField<string | undefined>("email", yup.string().email().required(), {
  initialValue: props.user?.email,
});

const { value: username,
  errorMessage: usernameError,
  resetField: resetUsername,
} = useField<string | undefined>("username", yup.string().required(), {
  initialValue: props.user?.username,
});

const {
  value: password,
  errorMessage: passwordError,
  resetField: resetPassword,
} = useField<string | undefined>("password", undefined, {
  initialValue: undefined,
});

const {
  value: isConfirmed,
  resetField: resetIsConfirmed,
} = useField<boolean | undefined>("isConfirmed", undefined, {
  initialValue: props.user?.status === "confirmed",
});

const resetFormFields = () => {
  resetName();
  resetEmail();
  resetUsername();
  resetPassword();
  resetIsConfirmed();
};

const togglePasswordVisibility = () => {
  showPassword.value = !showPassword.value;
};

const setMaxNamespaces = () => {
  maxNamespaces.value = disableNamespaceCreation.value ? 0 : maxNamespaces.value;
};

const { handleSubmit } = useForm<IUser>();

const handleErrors = (error: AxiosError) => {
  if (!error.response?.data) return;

  const errorFields = error.response.data as string[];
  errorFields.forEach((field) => {
    switch (field) {
      case "username":
        usernameError.value = "This username is invalid!";
        break;
      case "name":
        nameError.value = "This name is invalid!";
        break;
      case "email":
        emailError.value = "This email is invalid!";
        break;
      case "password":
        passwordError.value = "This password is invalid!";
        break;
      default:
        break;
    }
  });
};

const submitUser = async (isCreating: boolean, userData: Record<string, unknown>) => {
  try {
    const userStoreAction = isCreating ? userStore.addUser : userStore.put;
    await userStoreAction(userData);

    snackbar.showSuccess(`User ${isCreating ? "added" : "updated"} successfully.`);

    await userStore.refresh();
    showDialog.value = false;
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      handleErrors(error as AxiosError);
    }
    snackbar.showError("Failed to submit the user data.");
  }
};

const getStatus = () => {
  if (props.createUser) return undefined;

  if (canChangeStatus) {
    return isConfirmed.value ? "confirmed" : "not-confirmed";
  }

  return props.user?.status;
};

const prepareUserData = (): Record<string, unknown> => ({
  name: name.value,
  email: email.value,
  username: username.value,
  password: password.value || "",
  max_namespaces: changeNamespaceLimit.value ? maxNamespaces.value : undefined,
  confirmed: !props.createUser ? isConfirmed.value : undefined,
  status: getStatus(),
  id: !props.createUser ? props.user?.id : undefined,
});

const validateErrors = (): boolean => !nameError.value && !emailError.value && !usernameError.value;

const onSubmit = handleSubmit(async () => {
  if (validateErrors()) {
    const userData = prepareUserData();
    await submitUser(!!props.createUser, userData);
  } else {
    snackbar.showError("Please fill in all required fields.");
  }
});

const close = () => {
  showDialog.value = false;
  resetFormFields();
};

watch(changeNamespaceLimit, (newValue) => {
  if (!newValue) disableNamespaceCreation.value = false;
});

watch(disableNamespaceCreation, (newValue) => {
  if (!newValue) maxNamespaces.value = 1;
});

defineExpose({
  showDialog,
  password,
  name,
  email,
  username,
  isConfirmed,
});
</script>
