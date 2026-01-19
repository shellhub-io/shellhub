<template>
  <v-btn
    v-if="createUser"
    class="mr-2"
    color="primary"
    tabindex="0"
    data-test="user-add-btn"
    text="Add User"
    @click="showDialog = true"
  />

  <v-tooltip
    v-else
    bottom
    anchor="bottom"
  >
    <template #activator="{ props }">
      <v-icon
        data-test="user-edit-btn"
        tag="button"
        dark
        v-bind="props"
        tabindex="0"
        icon="mdi-pencil"
        @click="showDialog = true"
      />
    </template>
    <span>Edit</span>
  </v-tooltip>

  <FormDialog
    v-model="showDialog"
    :title="`${createUser ? 'Add new' : 'Edit'} user`"
    icon="mdi-account"
    icon-color="primary"
    :confirm-text="createUser ? 'Create' : 'Update'"
    cancel-text="Cancel"
    @confirm="submitForm"
    @cancel="close"
    @close="close"
  >
    <v-card-text class="pa-6">
      <v-text-field
        v-model="name"
        data-test="name-field"
        label="Name"
        required
        :error-messages="nameError"
        color="primary"
      />
      <v-text-field
        v-model="username"
        data-test="username-field"
        label="Username"
        required
        :error-messages="usernameError"
        color="primary"
      />
      <v-text-field
        v-model="email"
        data-test="email-field"
        label="Email"
        required
        name="email"
        :error-messages="emailError"
        color="primary"
      />
      <v-text-field
        v-model="password"
        data-test="password-field"
        label="Password"
        :required="createUser"
        name="password"
        :error-messages="passwordError"
        color="primary"
        :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
        :type="showPassword ? 'text' : 'password'"
        @click:append-inner="togglePasswordVisibility"
      />
      <v-checkbox
        v-model="changeNamespaceLimit"
        data-test="change-namespace-limit-checkbox"
        label="Change the namespace creation limit for this user"
        color="primary"
        density="compact"
        hide-details
        @update:model-value="disableNamespaceCreation = false"
      />
      <v-checkbox
        v-if="changeNamespaceLimit"
        v-model="disableNamespaceCreation"
        data-test="disable-namespace-creation-checkbox"
        label="Disable namespace creation"
        color="primary"
        class="mb-3"
        density="compact"
        hide-details
        @update:model-value="setMaxNamespaces"
      />
      <v-number-input
        v-if="changeNamespaceLimit"
        v-model="maxNamespaces"
        data-test="max-namespaces-input"
        :disabled="disableNamespaceCreation"
        label="Namespace limit"
        :min="1"
        color="primary"
        variant="outlined"
      />
      <v-tooltip
        location="bottom start"
        class="text-center"
        :disabled="canChangeStatus"
        :text="statusTooltipMessage"
      >
        <template #activator="{ props }">
          <div v-bind="props">
            <v-checkbox
              v-if="!createUser"
              v-model="isConfirmed"
              data-test="is-confirmed-checkbox"
              label="User confirmed"
              :disabled="!canChangeStatus"
              density="compact"
              hide-details
              color="primary"
            />
          </div>
        </template>
      </v-tooltip>
      <v-tooltip
        location="bottom start"
        class="text-center"
        :disabled="canChangeAdmin"
        text="You cannot remove your own admin privileges"
      >
        <template #activator="{ props }">
          <div v-bind="props">
            <v-checkbox
              v-model="isAdmin"
              data-test="is-admin-checkbox"
              label="Admin user"
              :disabled="!canChangeAdmin"
              density="compact"
              hide-details
              color="primary"
            />
          </div>
        </template>
      </v-tooltip>
    </v-card-text>
  </FormDialog>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import axios, { AxiosError } from "axios";
import * as yup from "yup";
import { useField, useForm } from "vee-validate";
import useUsersStore from "@admin/store/modules/users";
import useAuthStore from "@admin/store/modules/auth";
import { IAdminUser, IAdminUserFormData } from "@admin/interfaces/IUser";
import useSnackbar from "@/helpers/snackbar";
import FormDialog from "@/components/Dialogs/FormDialog.vue";

const props = defineProps<{
  createUser?: boolean;
  user?: IAdminUser;
}>();

const showDialog = ref(false);
const showPassword = ref(false);
const changeNamespaceLimit = ref(props.user?.max_namespaces !== -1);
const disableNamespaceCreation = ref(props.user?.max_namespaces === 0);
const maxNamespaces = ref(props.user?.max_namespaces || 0);
const canChangeStatus = props.user?.status === "not-confirmed"; // Only allow changing status if the user is not confirmed
const snackbar = useSnackbar();
const usersStore = useUsersStore();
const authStore = useAuthStore();

const statusTooltipMessage = props.user?.status === "invited"
  ? "You cannot change the status of an invited user."
  : "You cannot remove confirmation from an user.";

const isCurrentUser = computed(() => props.user?.username === authStore.currentUser);
const canChangeAdmin = computed(() => !(isCurrentUser.value && props.user?.admin));

const {
  value: name,
  errorMessage: nameError,
  setErrors: setNameError,
  resetField: resetName,
} = useField<string | undefined>("name", yup.string().required(), {
  initialValue: props.user?.name,
});

const {
  value: email,
  errorMessage: emailError,
  setErrors: setEmailError,
  resetField: resetEmail,
} = useField<string | undefined>("email", yup.string().email().required(), {
  initialValue: props.user?.email,
});

const {
  value: username,
  errorMessage: usernameError,
  setErrors: setUsernameError,
  resetField: resetUsername,
} = useField<string | undefined>("username", yup.string().required(), {
  initialValue: props.user?.username,
});

const {
  value: password,
  errorMessage: passwordError,
  setErrors: setPasswordError,
  resetField: resetPassword,
} = useField<string | undefined>("password", undefined, {
  initialValue: undefined,
});

const { value: isConfirmed, resetField: resetIsConfirmed } = useField<
  boolean | undefined
>("isConfirmed", undefined, {
  initialValue: props.user?.status === "confirmed",
});

const { value: isAdmin, resetField: resetIsAdmin } = useField<
  boolean | undefined
>("isAdmin", undefined, {
  initialValue: props.user?.admin || false,
});

const resetFormFields = () => {
  resetName();
  resetEmail();
  resetUsername();
  resetPassword();
  resetIsConfirmed();
  resetIsAdmin();
};

const togglePasswordVisibility = () => {
  showPassword.value = !showPassword.value;
};

const setMaxNamespaces = () => {
  maxNamespaces.value = disableNamespaceCreation.value ? 0 : 1;
};

const { handleSubmit } = useForm<IAdminUser>();

const handleErrors = (error: AxiosError) => {
  if (!error.response?.data) return;
  const errorFields = error.response.data as string[];

  errorFields.forEach((field) => {
    switch (field) {
      case "username":
        setUsernameError("This username is invalid!");
        break;
      case "name":
        setNameError("This name is invalid!");
        break;
      case "email":
        setEmailError("This email is invalid!");
        break;
      case "password":
        setPasswordError("This password is invalid!");
        break;
      default:
        break;
    }
  });
};

const close = () => {
  showDialog.value = false;
  resetFormFields();
};

const submitUser = async (
  isCreating: boolean,
  userData: IAdminUserFormData,
) => {
  try {
    const usersStoreAction = isCreating
      ? usersStore.addUser
      : usersStore.updateUser;
    await usersStoreAction(userData);

    snackbar.showSuccess(
      `User ${isCreating ? "added" : "updated"} successfully.`,
    );

    await usersStore.fetchUsersList();
    close();
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

const prepareUserData = () =>
  ({
    name: name.value,
    email: email.value,
    username: username.value,
    password: password.value || "",
    max_namespaces: changeNamespaceLimit.value
      ? maxNamespaces.value
      : undefined,
    status: getStatus(),
    id: !props.createUser ? props.user?.id : undefined,
    admin: isAdmin.value,
  }) as IAdminUserFormData;

const validateErrors = (): boolean =>
  !nameError.value && !emailError.value && !usernameError.value;

const submitForm = handleSubmit(async () => {
  if (validateErrors()) {
    const userData = prepareUserData();
    await submitUser(!!props.createUser, userData);
  } else {
    snackbar.showError("Please fill in all required fields.");
  }
});
</script>
