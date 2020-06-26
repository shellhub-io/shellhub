import { required, integer } from 'vee-validate/dist/rules';
import { extend } from 'vee-validate';

extend('required', {
  ...required,
  message: 'This field is required',
});

extend('integer', {
  ...integer,
  message: 'This value must be a integer number',
});
