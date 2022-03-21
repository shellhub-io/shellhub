import {
  required,
  integer,
  email,
  confirmed,
} from 'vee-validate/dist/rules';
import { extend } from 'vee-validate';
import isValidHostname from 'is-valid-hostname';

import { parsePrivateKey, parseKey } from '@/sshpk';

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

extend('noDot', {
  validate: (value) => !/\./.test(value),
  message: 'The name must not contain dots',
});

extend('rfc1123', {
  validate: (value) => isValidHostname(value),
  message: 'You entered an invalid RFC1123 name',
});

extend('routeIdentifier', {
  validate: (value) => !/\/|@|&|:/.test(value),
  message: 'The name must not contain /, @, &, and :.',
});

extend('password', (value) => {
  if (value.length < 5 || value.length > 30) {
    return 'Your password should be 5-30 characters long';
  }
  return true;
});

extend('namespace', (value) => {
  if (value.length < 3 || value.length > 30) {
    return 'Your namespace should be 3-30 characters long';
  }
  return true;
});

extend('device', (value) => {
  if (value.length < 3 || value.length > 30) {
    return 'Your hostname should be 3-30 characters long';
  }
  return true;
});

extend('tag', (value) => {
  if (value.length < 3 || value.length > 255) {
    return 'Your tag should be 3-255 characters long';
  }
  return true;
});

extend('tagRequired', {
  ...required,
  message: 'You must choose at least one tag',
});

extend('tagsLength', {
  validate(value) {
    if (value.length > 3) {
      return 'The maximum capacity has reached';
    }
    return true;
  },
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

extend('parseKey', {
  validate(value, args) {
    try {
      if (args.typeKey === 'private') {
        parsePrivateKey(value);
      } else {
        parseKey(value);
      }

      return true;
    } catch (err) {
      return false;
    }
  },
  params: ['typeKey'],
  message: 'Not valid key',
});
