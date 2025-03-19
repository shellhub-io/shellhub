import moment from "moment";

export const formatFullDateTime = (date: string) => {
  if (date) {
    return moment(date).format("dddd, MMMM Do YYYY, h:mm:ss a");
  }
  return null;
};

export const formatShortDateTime = (date: string) => {
  if (date) {
    return moment(date).format("LLL");
  }
  return null;
};

export const formatDateOnly = (date : number) => {
  if (date) {
    return moment.unix(date).format("MMM Do YYYY");
  }
  return null;
};

export const formatAbbreviatedDateTime = (date : string) => {
  if (date) {
    return moment(date).format("ddd, MMM Do YY, h:mm:ss a");
  }
  return null;
};

export const getTimeFromNow = (date: string) => {
  if (date) {
    return moment(date).fromNow();
  }
  return null;
};
