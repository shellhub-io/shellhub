<template>
  <v-dialog
    v-model="localDialog"
    @click:outside="handleClose"
    max-width="400"
    v-bind="$attrs"
  >
    <v-card data-test="connector-form-card" class="bg-v-theme-surface">
      <div>
        <v-card-title class="text-headline bg-primary">
          {{ isEditing ? 'Edit Docker Connector' : 'New Docker Connector' }}
        </v-card-title>

        <v-card-text>

          <v-container>
            <v-row>
              <v-col sm="8" class="pb-0">
                <v-text-field
                  class="mt-1"
                  v-model="address"
                  label="Address"
                  :error-messages="addressError"
                  required
                  variant="underlined"
                  data-test="address-text"
                />
              </v-col>
              <p class="mt-7 pa-0"> : </p>
              <v-col class="pb-0">
                <v-text-field
                  v-model.number="port"
                  label="Port"
                  :error-messages="portError"
                  required
                  :rules="validatePort"
                  variant="outlined"
                  data-test="port-text"
                />
              </v-col>
            </v-row>
            <v-row>
              <v-col class="pa-0">
                <v-checkbox
                  v-model="isSecure"
                  label="Secure"
                  hint="Secure your connection using TLS Certificates,
                 find out more information about setting a Docker Environment
                with TLS in the"
                  persistent-hint
                  :color="isSecure ? 'primary' : ''">
                  <template v-slot:message="{ message }">
                    <div
                    >
                      {{ message }}
                      <span
                      ><a
                        href="https://docs.docker.com/engine/security/protect-access/#use-tls-https-to-protect-the-docker-daemon-socket"
                        target="_blank"
                        rel="noopener"
                      >Docker Documentation</a>
                      </span>
                    </div>
                  </template>
                </v-checkbox>
              </v-col>
            </v-row>
            <v-row v-if="isSecure">
              <v-col>
                <v-file-input
                  label="TLS Ca Certificate"
                  variant="underlined"
                  v-model="caCertificate"
                  @change="handleCertificateChange('ca')"
                  :error-messages="caCertificateError"
                />
                <v-file-input
                  label="TLS Certificate"
                  variant="underlined"
                  v-model="certificate"
                  @change="handleCertificateChange('cert')"
                  :error-messages="certificateError"
                />
                <v-file-input
                  label="TLS Key"
                  variant="underlined"
                  v-model="key"
                  @change="handleCertificateChange('key')"
                  :error-messages="keyError"
                  hint="Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats."
                  persistent-hint
                />
              </v-col>
            </v-row>
          </v-container>
        </v-card-text>
      </div>
      <v-card-actions>
        <v-spacer />
        <v-btn color="primary" data-test="close-btn" @click="handleClose"> Close </v-btn>
        <v-btn :disabled="hasError" color="primary" data-test="save-btn" @click="saveConnector">
          {{ isEditing ? 'Save' : 'Add' }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { useField } from "vee-validate";
import * as yup from "yup";
import { computed, ref, watch } from "vue";
import { envVariables } from "../../envVariables";
import { useStore } from "../../store";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import { IConnectorPayload } from "@/interfaces/IConnector";
import { parseCertificate, parsePrivateKeySsh } from "../../utils/validate";
import hasPermission from "../../utils/permission";
import { actions, authorizer } from "../../authorizer";
import handleError from "../../utils/handleError";

const props = defineProps({
  isEditing: {
    type: Boolean,
    required: true,
  },
  initialAddress: {
    type: String,
    required: false,
    default: "",
  },
  initialPort: {
    type: Number,
    required: false,
    default: 2375,
  },
  uid: {
    type: String,
    required: false,
    default: "",
  },
  initialSecure: {
    type: Boolean,
    required: false,
    default: false,
  },
  storeMethod: {
    type: Function,
    required: true,
  },
  showDialog: {
    type: Boolean,
    required: false,
    default: false,
  },
});

const emit = defineEmits(["update", "close"]);

const localDialog = ref(props.showDialog);

watch(() => props.showDialog, (newValue) => {
  localDialog.value = newValue;
});

const store = useStore();

// eslint-disable-next-line vue/max-len
const ipAddressRegex = /^(25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})\.(25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})\.(25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})\.(25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})$/;

const hasAuthorizationAdd = () => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(
      authorizer.role[role],
      actions.connector.add,
    );
  }
  return false;
};

const {
  value: address,
  errorMessage: addressError,
} = useField<string>(
  "address",
  yup
    .string()
    .required()
    .matches(ipAddressRegex, "Invalid IP address format"),
  {
    initialValue: props.initialAddress,
  },
);

const {
  value: port,
  errorMessage: portError,
} = useField<number>(
  "port",
  yup
    .number()
    .integer()
    .max(65535)
    .required(),
  {
    initialValue: props.initialPort,
  },
);

const validatePort = ref([(v: string) => Number.isInteger(Number(v)) || "The value must be an integer number"]);

const isSecure = ref(props.initialSecure);

watch([isSecure, port], ([newSecure, newPort]) => {
  if (newSecure && newPort === 2375) {
    port.value = 2376;
  } else if (!newSecure && newPort === 2376) {
    port.value = 2375;
  }
});

const caCertificate = ref<File>();
const caCertificateError = ref<string>("");

const certificate = ref<File>();
const certificateError = ref<string>("");

const key = ref<File>();
const keyError = ref<string>("");

// eslint-disable-next-line vue/max-len
const hasError = computed(() => !!addressError.value || !!portError.value || !!keyError.value || !!certificateError.value || !!caCertificateError.value || !hasAuthorizationAdd() || (isSecure.value && (!caCertificate.value || !certificate.value || !key.value)));

const readFile = async (file) => new Promise((resolve, reject) => {
  const reader = new FileReader();

  reader.onload = () => {
    resolve(reader.result);
  };

  reader.onerror = () => {
    reject(reader.error);
  };

  reader.readAsText(file);
});

const validateFile = async (certificate, type) => {
  try {
    const content = await readFile(certificate);
    switch (type) {
      case "key":
        parsePrivateKeySsh(content);
        break;
      case "ca":
        parseCertificate(content);
        break;
      case "cert":
        parseCertificate(content);
        break;
      default:
        break;
    }
  } catch (error) {
    const typedErr = error as { name: string };
    switch (typedErr.name) {
      case "CertificateParseError":
        if (type === "ca") {
          caCertificateError.value = "Error parsing the Certificate.";
          return;
        }
        certificateError.value = "Error trying to parse the Certificate, please try again.";
        break;
      case "KeyEncryptedError":
        keyError.value = "Encrypted Keys with Passphrases are not allowed, please use a Key without Passphrase";
        break;
      case "KeyParseError":
        keyError.value = "Error parsing the key, please try again.";
        break;
      default:
        break;
    }
  }
};

const handleCertificateChange = async (type) => {
  let file;

  switch (type) {
    case "ca":
      file = caCertificate.value;
      break;
    case "cert":
      file = certificate.value;
      break;
    case "key":
      file = key.value;
      break;
    default:
      return;
  }

  if (file) {
    // Only clear the error message for the specific type being changed
    switch (type) {
      case "ca":
        caCertificateError.value = "";
        break;
      case "cert":
        certificateError.value = "";
        break;
      case "key":
        keyError.value = "";
        break;
      default:
        break;
    }
    await validateFile(file, type);
  }
};

const saveConnector = async () => {
  const payload: IConnectorPayload = {
    enable: true,
    secure: isSecure.value,
    address: address.value,
    port: port.value,
  };

  if (props.uid) {
    payload.uid = props.uid;
  }

  if (isSecure.value) {
    payload.tls = {
      ca: await readFile(caCertificate.value),
      cert: await readFile(certificate.value),
      key: await readFile(key.value),
    };
  }

  if (envVariables.isCommunity) {
    store.commit("users/setShowPaywall", true);
    return;
  }
  try {
    await props.storeMethod(payload);
    store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      props.isEditing ? INotificationsSuccess.connectorEdit : INotificationsSuccess.connectorAdd,
    );
    emit("update");
    emit("close");
    localDialog.value = false;
  } catch (error) {
    store.dispatch(
      "snackbar/showSnackbarErrorAction",
      props.isEditing ? INotificationsError.connectorEdit : INotificationsError.connectorAdd,
    );
    handleError(error);
  }
};

const handleClose = () => {
  emit("close");
  localDialog.value = false;
};

defineExpose({ localDialog, isSecure });
</script>
