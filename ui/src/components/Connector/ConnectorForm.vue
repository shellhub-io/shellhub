<template>
  <BaseDialog
    v-model="showDialog"
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
                  hint="Supports RSA, DSA, ECDSA (NIST P-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats."
                  persistent-hint
                />
              </v-col>
            </v-row>
          </v-container>
        </v-card-text>
      </div>
      <v-card-actions>
        <v-spacer />
        <v-btn data-test="close-btn" @click="handleClose"> Close </v-btn>
        <v-btn :disabled="hasError" color="primary" data-test="save-btn" @click="saveConnector">
          {{ isEditing ? 'Save' : 'Add' }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useField } from "vee-validate";
import * as yup from "yup";
import { computed, ref, watch } from "vue";
import { envVariables } from "@/envVariables";
import { IConnectorPayload } from "@/interfaces/IConnector";
import { parseCertificate, parsePrivateKey } from "@/utils/sshKeys";
import hasPermission from "@/utils/permission";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
import useUsersStore from "@/store/modules/users";

const props = defineProps<{
  isEditing: boolean;
  storeMethod:(payload: IConnectorPayload) => Promise<void>;
  initialAddress?: string;
  initialPort?: number;
  uid?: string;
  initialSecure?: boolean;
}>();

const usersStore = useUsersStore();
const showDialog = defineModel<boolean>({ default: false });
const emit = defineEmits(["update"]);
const snackbar = useSnackbar();

// eslint-disable-next-line vue/max-len
const ipAddressRegex = /^(25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})\.(25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})\.(25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})\.(25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})$/;

const canAddConnector = hasPermission("connector:add");

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
    initialValue: props.initialPort || 2375,
  },
);

const validatePort = ref([(v: string) => Number.isInteger(Number(v)) || "The value must be an integer number"]);

const isSecure = ref(props.initialSecure || false);

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
const hasError = computed(() => !!addressError.value || !!portError.value || !!keyError.value || !!certificateError.value || !!caCertificateError.value || !canAddConnector || (isSecure.value && (!caCertificate.value || !certificate.value || !key.value)));

const readFile = async (file) => new Promise<string>((resolve, reject) => {
  const reader = new FileReader();

  reader.onload = () => {
    resolve(reader.result as string);
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
        parsePrivateKey(content);
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
    uid: props.uid || "",
  };

  if (isSecure.value) {
    payload.tls = {
      ca: await readFile(caCertificate.value),
      cert: await readFile(certificate.value),
      key: await readFile(key.value),
    };
  }

  if (envVariables.isCommunity) {
    usersStore.showPaywall = true;
    return;
  }
  try {
    await props.storeMethod(payload);
    snackbar.showSuccess(props.isEditing ? "Connector edited successfully" : "Connector added successfully");
    emit("update");
    showDialog.value = false;
  } catch (error) {
    snackbar.showError(props.isEditing ? "Failed to edit connector" : "Failed to add connector");
    handleError(error);
  }
};

const handleClose = () => {
  showDialog.value = false;
};

defineExpose({ showDialog, isSecure });
</script>
