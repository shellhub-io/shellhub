<template>
  <v-row v-bind="$attrs">
    <v-col>
      <h3>Namespace</h3>
    </v-col>

    <v-spacer />

    <v-col md="auto" class="ml-auto">
      <v-tooltip
        location="bottom"
        class="text-center"
        :disabled="hasAuthorizationRenameNamespace()"
      >
        <template v-slot:activator="{ props }">
          <div v-bind="props">
            <v-btn
              :disabled="!hasAuthorizationRenameNamespace()"
              color="primary"
              @click="editNamespace"
            >
              Rename Namespace
            </v-btn>
          </div>
        </template>
        <span> You don't have this kind of authorization. </span>
      </v-tooltip>
    </v-col>
  </v-row>

  <div class="mt-4 mb-2">
    <v-text-field
      v-model="name"
      class="ml-3"
      label="Name"
      :error-messages="nameError"
      variant="underlined"
      required
      data-test="name-text"
    />
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, watch } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import { useStore } from "../../store";
import hasPermission from "../../utils/permission";
import { actions, authorizer } from "../../authorizer";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";

export default defineComponent({
  setup() {
    const store = useStore();

    const namespace = computed(() => store.getters["namespaces/get"]);
    const tenant = computed(() => store.getters["auth/tenant"]);

    const {
      value: name,
      errorMessage: nameError,
      setErrors: setNameError,
    } = useField<string>(
      "name",
      yup
        .string()
        .min(3, "Your namespace should be 3-30 characters long")
        .max(30, "Your namespace should be 3-30 characters long")
        .required()
        .matches(/^[^.]*$/, "The name must not contain dots"),
      {
        initialValue: "",
      },
    );

    watch(namespace, (ns) => {
      name.value = ns.name;
    });

    onMounted(() => {
      if (!store.getters["auth/isLoggedIn"]) return;
      store.dispatch("namespaces/get", tenant.value);
    });

    const hasAuthorizationRenameNamespace = () => {
      const role = store.getters["auth/role"];
      if (role !== "") {
        return hasPermission(authorizer.role[role], actions.namespace.rename);
      }

      return false;
    };

    const editNamespace = async () => {
      if (!nameError.value) {
        try {
          await store.dispatch("namespaces/put", {
            id: tenant.value,
            name: name.value,
          });
          await store.dispatch("namespaces/fetch", {
            page: 1,
            perPage: 10,
            filter: "",
          });
          await store.dispatch("namespaces/get", tenant.value);
          store.dispatch(
            "snackbar/showSnackbarSuccessAction",
            INotificationsSuccess.namespaceEdit,
          );
        } catch (error: any) {
          if (error.response.status === 400) {
            setNameError("This name is not valid");
          } else if (error.response.status === 409) {
            setNameError("name used already");
          } else {
            store.dispatch(
              "snackbar/showSnackbarErrorAction",
              INotificationsError.namespaceEdit,
            );
            throw new Error(error);
          }
        }
      }
    };

    return {
      name,
      nameError,
      namespace,
      tenant,
      setNameError,
      hasAuthorizationRenameNamespace,
      editNamespace,
    };
  },
});
</script>
