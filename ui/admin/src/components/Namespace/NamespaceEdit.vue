<template>
  <v-tooltip bottom anchor="bottom">
    <template v-slot:activator="{ props }">
      <v-icon
        @click="dialog = !dialog"
        tag="a"
        dark
        v-bind="props"
        tabindex="0"
        aria-label="Editar Namespace"
        @keypress.enter="dialog = !dialog"
      >mdi-pencil
      </v-icon>
    </template>
    <span>Edit</span>
  </v-tooltip>

  <v-dialog v-model="dialog" max-width="400" transition="dialog-bottom-transition">
    <v-card>
      <v-card-title class="text-h5 pb-2"> Edit Namespace </v-card-title>
      <v-divider />

      <form @submit.prevent="onSubmit">
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
                    data-test="name-text"
                  />

                  <v-text-field
                    v-model="numberDevices"
                    label="Maximum Devices"
                    required
                    type="number"
                    :min="-1"
                    :error-messages="numberDevicesError"
                    color="primary"
                    variant="underlined"
                    data-test="maxDevices-text"
                  />
                  <div class="d-flex align-center justify-center">
                    <span class="mr-4 text-body-1">Session record:</span>
                    <v-switch
                      v-model="sessionRecord"
                      :error-messages="sessionRecordError"
                      color="primary"
                      hide-details
                    />
                  </div>
                </v-container>
              </v-col>
            </v-row>
          </v-container>
        </v-card-text>
        <v-card-actions class="pa-4">
          <v-spacer />
          <v-btn class="mr-2" color="dark" @click="dialog = false" type="reset"> Cancel </v-btn>
          <v-btn color="dark" type="submit" class="mr-4"> Save </v-btn>
        </v-card-actions>
      </form>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { useField } from "vee-validate";
import { ref, PropType, onMounted, watch } from "vue";
import * as yup from "yup";
import useSnackbarStore from "@admin/store/modules/snackbar";
import useNamespacesStore from "@admin/store/modules/namespaces";
import { INamespace } from "../../interfaces/INamespace";
import { INotificationsSuccess } from "../../interfaces/INotifications";

const props = defineProps({
  namespace: {
    type: Object as PropType<INamespace>,
    required: true,
    default: Object,
  },
});

const snackbarStore = useSnackbarStore();
const namespacesStore = useNamespacesStore();

const dialog = ref(false);

const { value: name, errorMessage: nameError } = useField<string | undefined>(
  "name",
  yup.string().required(),
);

const { value: numberDevices, errorMessage: numberDevicesError } = useField<number | undefined>(
  "name",
  yup.number().required(),
);

const { value: sessionRecord, errorMessage: sessionRecordError } = useField<
      boolean | undefined
    >("name", yup.boolean());

onMounted(() => {
  name.value = props.namespace?.name;
  numberDevices.value = props.namespace?.max_devices;
  sessionRecord.value = props.namespace?.settings.session_record;
});

const hasErrors = () => {
  if (nameError.value || numberDevicesError.value || sessionRecordError.value) {
    return true;
  }

  return false;
};

const onSubmit = async () => {
  if (!hasErrors()) {
    await namespacesStore.put({
      ...props.namespace as INamespace,
      name: name.value as string,
      max_devices: numberDevices.value as number,
      settings: { session_record: sessionRecord.value },
    });
    await namespacesStore.refresh();
    snackbarStore.showSnackbarSuccessAction(INotificationsSuccess.namespaceEdit);
    dialog.value = false;
  } else {
    snackbarStore.showSnackbarErrorDefault();
  }
};

watch(dialog, () => {
  if (!dialog.value) {
    name.value = props.namespace?.name;
    numberDevices.value = props.namespace?.max_devices;
    sessionRecord.value = props.namespace?.settings.session_record;
  }
});
</script>
