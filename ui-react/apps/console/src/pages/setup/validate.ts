const USERNAME_REGEX = /^[a-z0-9\-_.@]+$/;

export interface FormErrors {
  name?: string;
  username?: string;
  email?: string;
  password?: string;
  confirmPassword?: string;
}

export function validate(fields: {
  name: string;
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
}): FormErrors {
  const errors: FormErrors = {};

  if (fields.name.length < 1 || fields.name.length > 64) {
    errors.name = "Name must be 1-64 characters long";
  }

  if (fields.username.length < 3 || fields.username.length > 32) {
    errors.username = "Username must be 3-32 characters long";
  } else if (!USERNAME_REGEX.test(fields.username)) {
    errors.username = "Only lowercase letters, numbers, and -_.@ are allowed";
  }

  if (!fields.email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(fields.email)) {
    errors.email = "Enter a valid email address";
  }

  if (fields.password.length < 5 || fields.password.length > 32) {
    errors.password = "Password must be 5-32 characters long";
  }

  if (fields.confirmPassword !== fields.password) {
    errors.confirmPassword = "Passwords do not match";
  }

  return errors;
}
