<template>
  <v-btn
    v-if="createUser"
    @click="openDialog"
    class="mr-2"
    outlined
    tabindex="0"
    aria-label="Dialog Add user"
    @keypress.enter="openDialog"
    data-test="user-add-btn"
  >
    Add User
  </v-btn>

  <v-tooltip v-else bottom anchor="bottom">
    <template v-slot:activator="{ props }">
      <v-icon
        @click="openDialog"
        tag="a"
        dark
        v-bind="props"
        tabindex="0"
        aria-label="Dialog edit user"
        @keypress.enter="openDialog"
      >mdi-pencil
      </v-icon>
    </template>
    <span>Edit</span>
  </v-tooltip>

  <v-dialog v-model="dialog" max-width="400" transition="dialog-bottom-transition">
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
                  <v-tooltip location="bottom" class="text-center" :disabled="!emailIsConfirmed">
                    <template v-slot:activator="{ props }">
                      <div v-bind="props">
                        <v-checkbox
                          v-if="!createUser"
                          label="User confirmed"
                          v-model="userConfirmed"
                          :error-messages="userConfirmedError"
                          :disabled="emailIsConfirmed"
                          density="compact"
                          hide-details
                          color="primary"
                        />
                      </div>
                    </template>
                    <span>You cannot unsubscribe the user's email confirmation</span>
                  </v-tooltip>
                </v-container>
              </v-col>
            </v-row>
          </v-container>
        </v-card-text>
        <v-card-actions class="pa-4">
          <v-spacer />
          <v-btn class="mr-2" color="dark" @click="dialog = false" type="reset">Cancel</v-btn>
          <v-btn class="mr-2" color="dark" type="submit">
            <span v-if="createUser">Create</span>
            <span v-else>Update</span>
          </v-btn>
        </v-card-actions>
      </form>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, PropType } from "vue";
import axios, { AxiosError } from "axios";
import * as yup from "yup";
import { useField, useForm } from "vee-validate";
import useSnackbarStore from "@admin/store/modules/snackbar";
import useUsersStore from "@admin/store/modules/users";
import { INotificationsSuccess } from "../../interfaces/INotifications";

type UserLocal = {
  id?: string;
  name?: string;
  email?: string;
  username?: string;
  password?: string;
  confirmed?: boolean;
  max_namespaces?: number;
};

const props = defineProps({
  createUser: {
    type: Boolean,
    default: false,
  },
  user: {
    type: Object as PropType<UserLocal>,
    default: () => ({}),
  },
  titleCard: {
    type: String,
    required: true,
  },
});

const dialog = ref(false);
const showPassword = ref(false);
const changeNamespaceLimit = ref(false);
const disableNamespaceCreation = ref(false);
const maxNamespaces = ref<number | undefined>(props.user?.max_namespaces || 0);
const emailIsConfirmed = computed(() => props.user?.confirmed);
const snackbarStore = useSnackbarStore();
const userStore = useUsersStore();

const { value: name,
  errorMessage: nameError,
  resetField: resetName,
} = useField<string | undefined>("name", yup.string().required());

const { value: email,
  errorMessage: emailError,
  resetField: resetEmail,
} = useField<string | undefined>("email", yup.string().email().required());

const { value: username,
  errorMessage: usernameError,
  resetField: resetUsername,
} = useField<string | undefined>("username", yup.string().required());

const {
  value: password,
  errorMessage: passwordError,
  resetField: resetPassword,
} = useField<string | undefined>("password");

const {
  value: userConfirmed,
  errorMessage: userConfirmedError,
  resetField: resetUserConfirmed,
} = useField<boolean | undefined>("userConfirmed");

const resetFormFields = () => {
  resetName();
  resetEmail();
  resetUsername();
  resetPassword();
  resetUserConfirmed();
};

const populateFieldsFromProps = () => {
  name.value = props.user?.name;
  email.value = props.user?.email;
  username.value = props.user?.username;
  password.value = undefined;
  userConfirmed.value = props.user?.confirmed;
  maxNamespaces.value = props.user?.max_namespaces || 0;
  changeNamespaceLimit.value = props.user?.max_namespaces !== -1;
};

const openDialog = () => {
  dialog.value = true;
  resetFormFields();
  if (!props.createUser) {
    populateFieldsFromProps();
  }
};

const togglePasswordVisibility = () => {
  showPassword.value = !showPassword.value;
};

const setMaxNamespaces = () => {
  maxNamespaces.value = disableNamespaceCreation.value ? 0 : maxNamespaces.value;
};

const { handleSubmit } = useForm<UserLocal>();

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
    if (isCreating) {
      await userStore.addUser(userData);
      snackbarStore.showSnackbarSuccessAction(INotificationsSuccess.addUser);
    } else {
      await userStore.put(userData);
      snackbarStore.showSnackbarSuccessAction(INotificationsSuccess.userEdit);
    }

    await userStore.refresh();
    dialog.value = false;
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      handleErrors(error as AxiosError);
      snackbarStore.showSnackbarErrorDefault();
    } else {
      snackbarStore.showSnackbarErrorDefault();
    }
  }
};

const prepareUserData = (): Record<string, unknown> => ({
  name: name.value,
  email: email.value,
  username: username.value,
  password: password.value || "",
  max_namespaces: changeNamespaceLimit.value ? maxNamespaces.value : undefined,
  confirmed: !props.createUser ? userConfirmed.value : undefined,
  id: !props.createUser ? props.user?.id : undefined,
});

const validateErrors = (): boolean => !nameError.value && !emailError.value && !usernameError.value;

const onSubmit = handleSubmit(async () => {
  if (validateErrors()) {
    const userData = prepareUserData();
    await submitUser(!!props.createUser, userData);
  } else {
    snackbarStore.showSnackbarErrorDefault();
  }
});

watch(dialog, (newValue) => {
  if (!newValue) resetFormFields();
});

watch(changeNamespaceLimit, (newValue) => {
  if (!newValue) disableNamespaceCreation.value = false;
});

watch(disableNamespaceCreation, (newValue) => {
  if (!newValue) maxNamespaces.value = 1;
});

defineExpose({
  openDialog,
  emailIsConfirmed,
  userConfirmed,
  password,
  name,
  email,
  username,
});
</script>
