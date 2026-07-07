export interface SamlFormValues {
  useMetadataUrl: boolean;
  metadataUrl: string;
  postUrl: string;
  redirectUrl: string;
  entityId: string;
  certificate: string;
  emailMapping: string;
  nameMapping: string;
  signRequests: boolean;
}

type FieldErrors = Partial<Record<keyof SamlFormValues, string>>;

function isValidUrl(s: string): boolean {
  try {
    new URL(s);
    return true;
  } catch {
    return false;
  }
}

function isCertValid(s: string): boolean {
  return (
    s.includes("-----BEGIN CERTIFICATE-----") &&
    s.includes("-----END CERTIFICATE-----")
  );
}

export function samlResolver(values: SamlFormValues): FieldErrors {
  const errors: FieldErrors = {};

  if (values.useMetadataUrl) {
    if (!values.metadataUrl || !isValidUrl(values.metadataUrl)) {
      errors.metadataUrl = "Must be a valid URL";
    }
    return errors;
  }

  if (!values.postUrl && !values.redirectUrl) {
    errors.postUrl = "At least one Sign-On URL (POST or Redirect) is required";
  } else {
    if (values.postUrl && !isValidUrl(values.postUrl)) {
      errors.postUrl = "Must be a valid URL";
    }

    if (values.redirectUrl && !isValidUrl(values.redirectUrl)) {
      errors.redirectUrl = "Must be a valid URL";
    }
  }

  if (!values.entityId.trim()) {
    errors.entityId = "Entity ID is required";
  }

  if (!values.certificate.trim()) {
    errors.certificate = "Certificate is required";
  } else if (!isCertValid(values.certificate)) {
    errors.certificate =
      "Must include BEGIN CERTIFICATE and END CERTIFICATE markers";
  }

  return errors;
}
