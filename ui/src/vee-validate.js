import { required, integer } from 'vee-validate/dist/rules';
import { extend } from 'vee-validate';
import isValidHostname from 'is-valid-hostname';

extend('required', {
  ...required,
  message: 'This field is required',
});

extend('integer', {
  ...integer,
  message: 'This value must be a integer number',
});

extend('rfc1123', {
  validate: (value) => isValidHostname(value),
  message: 'You entered an invalid RFC1123 hostname',
});
