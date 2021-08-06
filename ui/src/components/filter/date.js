import moment from 'moment';

export const formatDate = (date) => {
  if (date) {
    return moment(date).format('dddd, MMMM Do YYYY, h:mm:ss a');
  }
  return null;
};

export const formatDateWithoutDayAndHours = (date) => {
  if (date) {
    return moment(date).format('MMMM Do YYYY');
  }
  return null;
};

export const lastSeen = (date) => {
  if (date) {
    return moment(date).fromNow();
  }
  return null;
};
