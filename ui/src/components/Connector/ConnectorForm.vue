<template>
  <FormDialog
    v-model="showDialog"
    :title="isEditing ? 'Edit Docker Connector' : 'New Docker Connector'"
    icon="mdi-docker"
    icon-color="primary"
    :confirm-text="isEditing ? 'Save' : 'Add'"
    confirm-color="primary"
    cancel-text="Close"
    :confirm-disabled="hasError"
    confirm-data-test="save-btn"
    cancel-data-test="close-btn"
    footer-helper-text="Secure your connection using TLS Certificates, find out more information in the"
    footer-helper-link-text="Docker Documentation"
    footer-helper-link="https://docs.docker.com/engine/security/protect-access/#use-tls-https-to-protect-the-docker-daemon-socket"
    @confirm="saveConnector"
    @cancel="handleClose"
    @close="handleClose"
  >
    <v-card-text>
      <div class="d-flex align-center ga-2">
        <v-text-field
          v-model="address"
          width="60%"
          label="Address"
          :error-messages="addressError"
          required
          data-test="address-text"
        />
        <span class="mt-n4">:</span>
        <v-text-field
          v-model.number="port"
          label="Port"
          :error-messages="portError"
          required
          :rules="validatePort"
          data-test="port-text"
        />
      </div>
      <v-checkbox
        v-model="isSecure"
        label="Secure"
        density="compact"
        hint="Secure your connection using TLS Certificates"
        persistent-hint
        :color="isSecure ? 'primary' : ''"
      />
      <div
        v-if="isSecure"
        class="mt-2 py-2"
      >
        <v-file-input
          v-model="caCertificate"
          label="TLS Ca Certificate"
          :error-messages="caCertificateError"
          @change="handleCertificateChange('ca')"
        />
        <v-file-input
          v-model="certificate"
          label="TLS Certificate"
          :error-messages="certificateError"
          @change="handleCertificateChange('cert')"
        />
        <v-file-input
          v-model="key"
          label="TLS Key"
          :error-messages="keyError"
          hint="Supports RSA, DSA, ECDSA (NIST P-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats."
          persistent-hint
          @change="handleCertificateChange('key')"
        />
      </div>
    </v-card-text>
  </FormDialog>
</template>

<script setup lang="ts">
import { useField } from "vee-validate";
import * as yup from "yup";
import { computed, ref, watch } from "vue";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import { envVariables } from "@/envVariables";
import { IConnectorPayload } from "@/interfaces/IConnector";
import { parseCertificate, parsePrivateKey } from "@/utils/sshKeys";
import hasPermission from "@/utils/permission";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useUsersStore from "@/store/modules/users";

const props = defineProps<{
  isEditing: boolean;
  storeMethod: (payload: IConnectorPayload) => Promise<void>;
  initialAddress?: string;
  initialPort?: number;
  uid?: string;
  initialSecure?: boolean;
}>();

const usersStore = useUsersStore();
const showDialog = defineModel<boolean>({ required: true });
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

const readFile = async (file: File) => new Promise<string>((resolve, reject) => {
  const reader = new FileReader();

  reader.onload = () => {
    resolve(reader.result as string);
  };

  reader.onerror = () => {
    reject(new Error(reader.error?.message || "Failed to read file"));
  };

  reader.readAsText(file);
});

const validateFile = async (certificate: File, type: "key" | "ca" | "cert") => {
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

const handleCertificateChange = async (type: "key" | "ca" | "cert") => {
  let file: File;

  switch (type) {
    case "ca":
      file = caCertificate.value as File;
      break;
    case "cert":
      file = certificate.value as File;
      break;
    case "key":
      file = key.value as File;
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
      ca: await readFile(caCertificate.value as File),
      cert: await readFile(certificate.value as File),
      key: await readFile(key.value as File),
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
