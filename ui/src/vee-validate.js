import {
  required,
  integer,
  email,
  confirmed,
} from 'vee-validate/dist/rules';
import { extend } from 'vee-validate';
import isValidHostname from 'is-valid-hostname';

import { parseKey } from '@/sshpk';

extend('required', {
  ...required,
  message: 'This field is required',
});

extend('integer', {
  ...integer,
  message: 'This value must be a integer number',
});

extend('email', {
  ...email,
  message: 'This field must be a valid email',
});

extend('rfc1123', {
  validate: (value) => isValidHostname(value),
  message: 'You entered an invalid RFC1123 name',
});

extend('password', (value) => {
  if (value.length < 5 || value.length > 30) {
    return 'Your password should be 5-30 characters long';
  }
  return true;
});

extend('comparePasswords', {
  validate(value, { currentPassword }) {
    if (value === currentPassword) {
      return false;
    }
    return true;
  },
  params: ['currentPassword'],
  message: 'The passwords are the same',
});

extend('confirmed', {
  ...confirmed,
  message: 'The passwords do not match',
});

extend('parseKey', (value) => {
  try {
    parseKey(value);
    return true;
  } catch (err) {
    return 'It\'s not a valid';
  }
});
